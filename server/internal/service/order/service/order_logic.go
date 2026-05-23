package service

import (
	"context"
	"fmt"
	"time"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/order/model"
	"github.com/zhian9/GoForge/server/internal/service/order/repository"
	"github.com/zeromicro/go-zero/core/logx"
)

// OrderLogic 订单业务逻辑
type OrderLogic struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	orderLogRepo  repository.OrderLogRepository
	cache         *cache.CacheOperations
	mqProducer    *mq.Producer
}

// NewOrderLogic 创建订单业务逻辑
func NewOrderLogic(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	orderLogRepo repository.OrderLogRepository,
	cache *cache.CacheOperations,
	mqProducer *mq.Producer,
) *OrderLogic {
	return &OrderLogic{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		orderLogRepo:  orderLogRepo,
		cache:         cache,
		mqProducer:    mqProducer,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID          uint64
	AddressID       uint64
	Items           []OrderItemRequest
	OrderType       int8
	CouponID        uint64
	Remark          string
	ReceiverName    string
	ReceiverPhone   string
	ReceiverAddress string
	ClearCart       bool
}

// OrderItemRequest 订单商品项请求
type OrderItemRequest struct {
	SkuID       uint64
	Quantity    int
	ProductName string
	Price       float64
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	Order *model.Order
}

// CreateOrder 创建订单
// 注意：这里简化了实现，实际应该调用商品服务获取SKU信息，调用库存服务扣减库存等
func (l *OrderLogic) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 1. 参数验证
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}
	if req.AddressID == 0 {
		return nil, apperrors.NewInvalidParamError("收货地址ID不能为空")
	}
	if len(req.Items) == 0 {
		return nil, apperrors.NewInvalidParamError("订单商品项不能为空")
	}

	// 2. 生成订单号（使用Redis保证唯一性）
	orderNo := l.generateOrderNo(ctx)

	// 3. 创建订单（简化实现，实际需要调用其他服务获取商品信息、地址信息等）
	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          req.UserID,
		OrderType:       req.OrderType,
		Status:          model.OrderStatusPending,
		TotalAmount:     0,
		PayAmount:       0, // 需要计算
		DiscountAmount:  0,
		FreightAmount:   0,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
	}

	if req.Remark != "" {
		order.Remark = &req.Remark
	}

	// 4. 保存订单
	if err := l.orderRepo.Create(ctx, order); err != nil {
		return nil, apperrors.NewInternalError("创建订单失败: " + err.Error())
	}

	// 新订单创建后，清理该用户订单列表缓存，避免列表不刷新
	if l.cache != nil {
		_ = l.cache.DeletePattern(ctx, fmt.Sprintf("%s%d:*", cache.KeyPrefixOrderList, order.UserID))
	}

	// 5. 创建订单商品项（简化实现）
	items := make([]*model.OrderItem, 0, len(req.Items))
	for _, itemReq := range req.Items {
		item := &model.OrderItem{
			OrderID:     order.ID,
			OrderNo:     orderNo,
			ProductID:   0,  // 需要从商品服务获取
			ProductName: itemReq.ProductName, // 需要从商品服务获取
			SkuID:       itemReq.SkuID,
			SkuCode:     "", // 需要从商品服务获取
			SkuName:     "", // 需要从商品服务获取
			Price:       itemReq.Price,  // 需要从商品服务获取
			Quantity:    itemReq.Quantity,
			TotalAmount: itemReq.Price * float64(itemReq.Quantity), // 需要计算
		}
		items = append(items, item)
	}

	if err := l.orderItemRepo.CreateBatch(ctx, items); err != nil {
		return nil, apperrors.NewInternalError("创建订单商品项失败: " + err.Error())
	}

	// 6. 记录订单日志
	log := &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      orderNo,
		OperatorType: 1, // 用户
		OperatorID:   &req.UserID,
		Action:       "create",
		AfterStatus:  &order.Status,
		Remark:       &req.Remark,
	}
	if err := l.orderLogRepo.Create(ctx, log); err != nil {
		// 日志记录失败不影响主流程
	}

	// 7. 发送订单创建Kafka消息
	if l.mqProducer != nil {
		message := mq.NewMessage(mq.TopicOrderCreated, map[string]interface{}{
			"order_id":     order.ID,
			"order_no":     orderNo,
			"user_id":      req.UserID,
			"total_amount": order.TotalAmount,
			"created_at":   time.Now().Format(time.RFC3339),
		})
		_ = l.mqProducer.PublishWithKey(ctx, mq.TopicOrderCreated, orderNo, message)
	}

	// 8. 清空购物车（结算完成后清除已选商品）
	if req.ClearCart && l.cache != nil {
		cartKey := fmt.Sprintf("cart-service:user:%d", req.UserID)
		_ = l.cache.Delete(ctx, cartKey)
	}

	return &CreateOrderResponse{
		Order: order,
	}, nil
}

// GetOrderRequest 获取订单详情请求
type GetOrderRequest struct {
	ID      uint64
	OrderNo string
}

// GetOrderResponse 获取订单详情响应
type GetOrderResponse struct {
	Order      *model.Order
	OrderItems []*model.OrderItem
}

// GetOrder 获取订单详情（带缓存）
func (l *OrderLogic) GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	// 确定订单ID
	var orderID uint64
	var orderNo string
	var err error

	if req.ID > 0 {
		orderID = req.ID
	} else if req.OrderNo != "" {
		orderNo = req.OrderNo
		// 先通过订单号查询订单ID
		order, err := l.orderRepo.GetByOrderNo(ctx, orderNo)
		if err != nil {
			return nil, apperrors.NewInternalError("查询订单失败: " + err.Error())
		}
		if order == nil {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "订单不存在")
		}
		orderID = order.ID
	} else {
		return nil, apperrors.NewInvalidParamError("订单ID或订单号不能为空")
	}

	// 尝试从缓存获取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixOrderDetail, orderID)
		var cachedResp GetOrderResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// 从数据库查询订单
	order, err := l.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单失败: " + err.Error())
	}
	if order == nil {
		return nil, apperrors.NewError(apperrors.CodeOrderNotFound, "订单不存在")
	}

	// 获取订单商品项
	items, err := l.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单商品项失败: " + err.Error())
	}

	return &GetOrderResponse{
		Order:      order,
		OrderItems: items,
	}, nil
}

// ListOrdersRequest 获取订单列表请求
type ListOrdersRequest struct {
	UserID   uint64
	Status   int8
	Page     int
	PageSize int
}

// ListOrdersResponse 获取订单列表响应
type ListOrdersResponse struct {
	Orders     []*model.Order
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

// ListOrders 获取订单列表（带缓存）
func (l *OrderLogic) ListOrders(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error) {
	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建缓存键
	if l.cache != nil {
		cacheKey := fmt.Sprintf("%s%d:%d:%d",
			cache.KeyPrefixOrderList,
			req.UserID,
			req.Status,
			req.Page,
		)
		var cachedResp ListOrdersResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// 构建查询请求
	repoReq := &repository.ListOrdersRequest{
		UserID:   req.UserID,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询订单列表
	orders, total, err := l.orderRepo.List(ctx, repoReq)
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单列表失败: " + err.Error())
	}

	// 为每个订单加载订单项
	for _, order := range orders {
		items, err := l.orderItemRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			// 记录错误但不中断，因为订单本身是成功的
			logx.Errorf("加载订单项失败 order_id=%d: %v", order.ID, err)
			order.Items = []model.OrderItem{} // 设置为空数组
		} else {
			// 转换 []*model.OrderItem 为 []model.OrderItem
			orderItems := make([]model.OrderItem, len(items))
			for i, item := range items {
				if item != nil {
					orderItems[i] = *item
				}
			}
			order.Items = orderItems
		}
	}

	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	resp := &ListOrdersResponse{
		Orders:     orders,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	// 写入缓存（5分钟）
	if l.cache != nil {
		cacheKey := fmt.Sprintf("%s%d:%d:%d",
			cache.KeyPrefixOrderList,
			req.UserID,
			req.Status,
			req.Page,
		)
		_ = l.cache.Set(ctx, cacheKey, resp, 5*time.Minute)
	}

	return resp, nil
}

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	ID      uint64
	OrderNo string
	Reason  string
}

// CancelOrderResponse 取消订单响应
type CancelOrderResponse struct {
	Success bool
}

// CancelOrder 取消订单
func (l *OrderLogic) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*CancelOrderResponse, error) {
	// 获取订单
	getReq := &GetOrderRequest{
		ID:      req.ID,
		OrderNo: req.OrderNo,
	}
	getResp, err := l.GetOrder(ctx, getReq)
	if err != nil {
		return nil, err
	}

	order := getResp.Order

	// 检查订单状态
	if order.Status != model.OrderStatusPending {
		return nil, apperrors.NewError(apperrors.CodeForbidden, "只能取消待支付订单")
	}

	// 更新订单状态
	reason := req.Reason
	if err := l.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusCancelled, &reason); err != nil {
		return nil, apperrors.NewInternalError("取消订单失败: " + err.Error())
	}

	// 删除缓存
	if l.cache != nil {
		detailKey := cache.BuildKey(cache.KeyPrefixOrderDetail, order.ID)
		_ = l.cache.Delete(ctx, detailKey)
		// 删除该用户所有订单列表缓存（避免取消后列表还是“待支付”，导致再次取消报 403）
		_ = l.cache.DeletePattern(ctx, fmt.Sprintf("%s%d:*", cache.KeyPrefixOrderList, order.UserID))
	}

	// 记录订单日志
	log := &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		OperatorType: 1, // 用户
		Action:       "cancel",
		BeforeStatus: &order.Status,
		AfterStatus:  func() *int8 { s := model.OrderStatusCancelled; return &s }(),
		Remark:       &reason,
	}
	_ = l.orderLogRepo.Create(ctx, log)

	// 发送订单取消Kafka消息
	if l.mqProducer != nil {
		message := mq.NewMessage(mq.TopicOrderCancelled, map[string]interface{}{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"reason":   reason,
		})
		_ = l.mqProducer.PublishWithKey(ctx, mq.TopicOrderCancelled, order.OrderNo, message)
	}

	return &CancelOrderResponse{
		Success: true,
	}, nil
}

// ConfirmReceiveRequest 确认收货请求
type ConfirmReceiveRequest struct {
	ID      uint64
	OrderNo string
}

// ConfirmReceiveResponse 确认收货响应
type ConfirmReceiveResponse struct {
	Success bool
}

// ConfirmReceive 确认收货
func (l *OrderLogic) ConfirmReceive(ctx context.Context, req *ConfirmReceiveRequest) (*ConfirmReceiveResponse, error) {
	// 获取订单
	getReq := &GetOrderRequest{
		ID:      req.ID,
		OrderNo: req.OrderNo,
	}
	getResp, err := l.GetOrder(ctx, getReq)
	if err != nil {
		return nil, err
	}

	order := getResp.Order

	// 检查订单状态
	if order.Status != model.OrderStatusShipped {
		return nil, apperrors.NewError(apperrors.CodeForbidden, "只能确认已发货的订单")
	}

	// 更新订单状态
	if err := l.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusCompleted, nil); err != nil {
		return nil, apperrors.NewInternalError("确认收货失败: " + err.Error())
	}

	// 记录订单日志
	log := &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		OperatorType: 1, // 用户
		Action:       "confirm_receive",
		BeforeStatus: &order.Status,
		AfterStatus:  func() *int8 { s := model.OrderStatusCompleted; return &s }(),
	}
	_ = l.orderLogRepo.Create(ctx, log)

	return &ConfirmReceiveResponse{
		Success: true,
	}, nil
}

// generateOrderNo 生成订单号（使用Redis保证唯一性）
func (l *OrderLogic) generateOrderNo(ctx context.Context) string {
	if l.cache != nil {
		// 使用日期作为key，Redis INCR保证唯一性
		today := time.Now().Format("20060102")
		seqKey := cache.BuildKey(cache.KeyPrefixOrderSeq, today)
		seq, err := l.cache.Increment(ctx, seqKey)
		if err == nil {
			// 设置过期时间为24小时
			_ = l.cache.Expire(ctx, seqKey, 24*time.Hour)
			return fmt.Sprintf("ORD%s%06d", today, seq)
		}
	}
	// 如果Redis不可用，使用时间戳
	now := time.Now()
	return fmt.Sprintf("ORD%s%06d",
		now.Format("20060102"),
		now.Nanosecond()%1000000,
	)
}
