package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"github.com/zhian9/GoForge/server/internal/service/order/model"
	"github.com/zhian9/GoForge/server/internal/service/order/repository"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

// OrderLogic 订单业务逻辑
type OrderLogic struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	orderLogRepo  repository.OrderLogRepository
	cache         *cache.CacheOperations
	mqProducer    *mq.Producer
	db            *gorm.DB
}

// NewOrderLogic 创建订单业务逻辑
func NewOrderLogic(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	orderLogRepo repository.OrderLogRepository,
	cache *cache.CacheOperations,
	mqProducer *mq.Producer,
	db *gorm.DB,
) *OrderLogic {
	return &OrderLogic{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		orderLogRepo:  orderLogRepo,
		cache:         cache,
		mqProducer:    mqProducer,
		db:            db,
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

	// 3. 计算订单金额：前端未传价格时查 SKU 表，汇总总额
	type resolvedItem struct {
		SkuID       uint64
		Quantity    int
		ProductName string
		Price       float64
	}
	// 先收集缺价格的 SKU，一次 IN 查询取回，避免在下单循环里逐条查（N+1）
	missingPriceSkuIDs := make([]uint64, 0, len(req.Items))
	for _, itemReq := range req.Items {
		if itemReq.Price <= 0 {
			missingPriceSkuIDs = append(missingPriceSkuIDs, itemReq.SkuID)
		}
	}
	priceBySkuID := make(map[uint64]float64, len(missingPriceSkuIDs))
	if len(missingPriceSkuIDs) > 0 && l.db != nil {
		var rows []struct {
			ID    uint64  `gorm:"column:id"`
			Price float64 `gorm:"column:price"`
		}
		if err := l.db.WithContext(ctx).Table("sku").Select("id, price").Where("id IN ?", missingPriceSkuIDs).Scan(&rows).Error; err == nil {
			for _, row := range rows {
				if row.Price > 0 {
					priceBySkuID[row.ID] = row.Price
				}
			}
		}
	}
	resolvedItems := make([]resolvedItem, 0, len(req.Items))
	var totalAmount float64
	for _, itemReq := range req.Items {
		price := itemReq.Price
		if price <= 0 {
			price = priceBySkuID[itemReq.SkuID]
		}
		totalAmount += price * float64(itemReq.Quantity)
		resolvedItems = append(resolvedItems, resolvedItem{
			SkuID:       itemReq.SkuID,
			Quantity:    itemReq.Quantity,
			ProductName: itemReq.ProductName,
			Price:       price,
		})
	}

	// 4. 创建订单
	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          req.UserID,
		OrderType:       req.OrderType,
		Status:          model.OrderStatusPending,
		TotalAmount:     totalAmount,
		PayAmount:       totalAmount,
		DiscountAmount:  0,
		FreightAmount:   0,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
	}

	if req.Remark != "" {
		order.Remark = &req.Remark
	}

	// 5. 写订单 / 订单项 / 扣库存放在同一个事务里：
	//    任何一步失败都整体回滚，避免出现「订单已建但库存没扣」这类半截数据。
	if l.db == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化，无法创建订单")
	}
	txErr := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先锁定并试算优惠券（只认 token 里的 userID），再写订单，
		// 这样订单金额与券的核销状态在同一事务里保持一致。
		discount, err := lockAndCalcCoupon(ctx, tx, req.UserID, req.CouponID, totalAmount)
		if err != nil {
			return err
		}
		order.DiscountAmount = discount
		order.PayAmount = math.Round((totalAmount-discount)*100) / 100
		if order.PayAmount < 0 {
			order.PayAmount = 0
		}

		orderRepoTx := repository.NewOrderRepository(tx)
		if err := orderRepoTx.Create(ctx, order); err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 订单落库拿到 ID 之后再把券挂到订单上（status=0 的条件更新保证并发下只核销一次）
		if req.CouponID > 0 {
			if err := redeemUserCoupon(ctx, tx, req.CouponID, order.ID); err != nil {
				return err
			}
		}

		// 订单项依赖订单自增 ID，必须在订单写入之后再构造
		items := make([]*model.OrderItem, 0, len(resolvedItems))
		for _, it := range resolvedItems {
			items = append(items, &model.OrderItem{
				OrderID:     order.ID,
				OrderNo:     orderNo,
				ProductID:   0,              // 需要从商品服务获取
				ProductName: it.ProductName, // 需要从商品服务获取
				SkuID:       it.SkuID,
				SkuCode:     "", // 需要从商品服务获取
				SkuName:     "", // 需要从商品服务获取
				Price:       it.Price,
				Quantity:    it.Quantity,
				TotalAmount: it.Price * float64(it.Quantity),
			})
		}
		if len(items) > 0 {
			orderItemRepoTx := repository.NewOrderItemRepository(tx)
			if err := orderItemRepoTx.CreateBatch(ctx, items); err != nil {
				return fmt.Errorf("创建订单商品项失败: %w", err)
			}
		}

		// 扣减库存（sku.stock 为唯一真源，inventory 同步）
		for _, it := range resolvedItems {
			if err := deductStock(ctx, tx, l.cache, it.SkuID, it.Quantity); err != nil {
				return fmt.Errorf("扣减库存失败: %w", err)
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, apperrors.NewInternalError(txErr.Error())
	}

	// 事务提交后再失效缓存，避免回滚时把缓存清空却查不到新订单
	invalidateOrderListCache(ctx, l.cache, req.UserID)

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
	// 注意：必须带上 PageSize —— 否则「每页 6 条」和「每页 1 条」这类不同分页的请求
	// 会命中同一份缓存，出现「请求 page_size=3 却返回 page_size=6 的数据」这种串数据。
	if l.cache != nil {
		cacheKey := fmt.Sprintf("%s%d:%d:%d:%d",
			cache.KeyPrefixOrderList,
			req.UserID,
			req.Status,
			req.Page,
			req.PageSize,
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

	// 批量加载订单项：一次 IN 查询，避免每单一次的 N+1
	if len(orders) > 0 {
		orderIDs := make([]uint64, 0, len(orders))
		for _, order := range orders {
			orderIDs = append(orderIDs, order.ID)
		}
		itemsByOrderID, err := l.orderItemRepo.GetByOrderIDs(ctx, orderIDs)
		if err != nil {
			// 记录错误但不中断，因为订单本身是成功的
			logx.Errorf("批量加载订单项失败: %v", err)
		}
		for _, order := range orders {
			items := itemsByOrderID[order.ID]
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
		// key 必须和上面读取时完全一致（含 PageSize），否则缓存永远读不中
		cacheKey := fmt.Sprintf("%s%d:%d:%d:%d",
			cache.KeyPrefixOrderList,
			req.UserID,
			req.Status,
			req.Page,
			req.PageSize,
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

	// 回增库存：取消待支付订单时释放已扣减的 sku.stock / inventory
	items, err := l.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单商品项失败: " + err.Error())
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		if err := restockStock(ctx, l.db, l.cache, item.SkuID, item.Quantity); err != nil {
			return nil, apperrors.NewInternalError("回增库存失败: " + err.Error())
		}
	}

	// 退回优惠券（整单取消后券可再次使用）
	releaseUserCoupon(ctx, l.db, order.ID)

	// 删除缓存
	if l.cache != nil {
		detailKey := cache.BuildKey(cache.KeyPrefixOrderDetail, order.ID)
		_ = l.cache.Delete(ctx, detailKey)
		// 失效该用户订单列表缓存（避免取消后列表还是“待支付”，导致再次取消报 403）
		invalidateOrderListCache(ctx, l.cache, order.UserID)
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
			"user_id":  order.UserID,
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

// 物流公司编码 -> 名称映射（发货时使用）
var logisticsCompanyMap = map[string]string{
	"SF": "顺丰速运", "YTO": "圆通速递", "ZTO": "中通快递", "STO": "申通快递", "YD": "韵达快递",
	"JD": "京东物流", "YZ": "邮政EMS",
}

// jwtSecret 与网关/其他服务保持一致
const jwtSecret = "goforge-jwt-secret"

// CheckAdmin 供 gRPC handler 层复用的管理员校验（逻辑同 checkAdmin）
func (l *OrderLogic) CheckAdmin(ctx context.Context) error {
	return l.checkAdmin(ctx)
}

// checkAdmin 校验当前调用者是否为管理员（发货等管理操作使用）
func (l *OrderLogic) checkAdmin(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return apperrors.NewError(apperrors.CodeUnauthorized, "未授权")
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		authHeaders = md.Get("gateway-authorization")
	}
	if len(authHeaders) == 0 {
		return apperrors.NewError(apperrors.CodeUnauthorized, "请先登录")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeaders[0], "Bearer "))
	claims, err := utils.ParseToken(token, jwtSecret)
	if err != nil {
		return apperrors.NewError(apperrors.CodeUnauthorized, "登录已过期，请重新登录")
	}

	if l.db == nil {
		return apperrors.NewError(apperrors.CodeForbidden, "无权限操作")
	}
	var isAdmin int8
	if err := l.db.WithContext(ctx).Table("user").Select("is_admin").Where("id = ?", claims.UserID).Scan(&isAdmin).Error; err != nil {
		return apperrors.NewInternalError("校验权限失败")
	}
	if isAdmin != 1 {
		return apperrors.NewError(apperrors.CodeForbidden, "无权限操作，仅管理员可发货")
	}
	return nil
}

// ShipOrderRequest 发货请求
type ShipOrderRequest struct {
	ID          uint64
	CompanyCode string
	LogisticsNo string
}

// ShipOrderResponse 发货响应
type ShipOrderResponse struct {
	Success bool
}

// ShipOrder 发货：订单 待发货 -> 待收货，并创建物流单（直接写 logistics 表）
func (l *OrderLogic) ShipOrder(ctx context.Context, req *ShipOrderRequest) (*ShipOrderResponse, error) {
	// 管理员鉴权
	if err := l.checkAdmin(ctx); err != nil {
		return nil, err
	}

	order, err := l.orderRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单失败: " + err.Error())
	}
	if order == nil {
		return nil, apperrors.NewError(apperrors.CodeOrderNotFound, "订单不存在")
	}
	if order.Status != model.OrderStatusPaid {
		return nil, apperrors.NewError(apperrors.CodeForbidden, "只能对待发货订单执行发货")
	}

	companyCode := req.CompanyCode
	if companyCode == "" {
		companyCode = "EXP"
	}
	companyName := logisticsCompanyMap[companyCode]
	if companyName == "" {
		companyName = companyCode
	}
	logisticsNo := req.LogisticsNo
	if logisticsNo == "" {
		logisticsNo = fmt.Sprintf("%s%s%06d", companyCode, time.Now().Format("20060102150405"), rand.Intn(1000000))
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	// 创建物流单（已发货状态，轨迹首节点「已发货」）
	if l.db != nil {
		trackingJSON := fmt.Sprintf(`[{"time":"%s","status":"已发货","remark":"商品已发出"}]`, nowStr)
		insertErr := l.db.WithContext(ctx).Exec(
			"INSERT INTO logistics (order_id, order_no, logistics_company, logistics_no, receiver_name, receiver_phone, receiver_address, status, tracking_info, shipped_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?)",
			order.ID, order.OrderNo, companyName, logisticsNo, order.ReceiverName, order.ReceiverPhone, order.ReceiverAddress, trackingJSON, now, now, now,
		).Error
		if insertErr != nil {
			return nil, apperrors.NewInternalError("创建物流单失败: " + insertErr.Error())
		}
	}

	// 更新订单状态：待发货 -> 待收货
	if err := l.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusShipped, nil); err != nil {
		return nil, apperrors.NewInternalError("更新订单状态失败: " + err.Error())
	}

	// 清缓存
	if l.cache != nil {
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixOrderDetail, order.ID))
		invalidateOrderListCache(ctx, l.cache, order.UserID)
	}

	return &ShipOrderResponse{Success: true}, nil
}

// GetStatsRequest 统计请求
type GetStatsRequest struct{}

// GetStatsResponse 统计响应
type GetStatsResponse struct {
	TotalOrders int64
	TotalSales  float64
	TodayOrders int64
}

// GetStats 获取订单统计数据
func (l *OrderLogic) GetStats(ctx context.Context, req *GetStatsRequest) (*GetStatsResponse, error) {
	var totalOrders, todayOrders int64
	var totalSales float64
	if l.db != nil {
		_ = l.db.WithContext(ctx).Table("orders").Count(&totalOrders).Error
		_ = l.db.WithContext(ctx).Table("orders").
			Select("COALESCE(SUM(pay_amount), 0)").
			Where("status IN (2,3,4)").
			Scan(&totalSales).Error
		_ = l.db.WithContext(ctx).Table("orders").
			Where("created_at >= ?", time.Now().Format("2006-01-02")).
			Count(&todayOrders).Error
	}
	return &GetStatsResponse{TotalOrders: totalOrders, TotalSales: totalSales, TodayOrders: todayOrders}, nil
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

	// 清缓存（详情 + 列表），避免列表仍显示「待收货」
	if l.cache != nil {
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixOrderDetail, order.ID))
		invalidateOrderListCache(ctx, l.cache, order.UserID)
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

// deductStock 扣减库存：sku.stock 为唯一真源（原子 UPDATE 防超卖），inventory 表同步
func deductStock(ctx context.Context, db *gorm.DB, cacheOps *cache.CacheOperations, skuID uint64, quantity int) error {
	if db == nil {
		return nil
	}
	if quantity <= 0 {
		quantity = 1
	}
	// 原子扣减 sku.stock，stock >= quantity 防超卖
	res := db.WithContext(ctx).Exec(
		"UPDATE sku SET stock = stock - ? WHERE id = ? AND stock >= ?",
		quantity, skuID, quantity,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("SKU %d 库存不足", skuID)
	}
	// 同步 inventory：可用库存减少、已售增加
	_ = db.WithContext(ctx).Exec(
		"UPDATE inventory SET available_stock = available_stock - ?, sold_stock = sold_stock + ? WHERE sku_id = ?",
		quantity, quantity, skuID,
	).Error
	// 失效商品缓存：列表 / 详情 / SKU 信息，避免管理端商品列表库存显示旧值
	invalidateProductCache(ctx, db, cacheOps, skuID)
	return nil
}

// userCouponRow / couponRow 只取核销需要的字段（列名与 database/schema.sql 一致）
type userCouponRow struct {
	ID       uint64    `gorm:"column:id"`
	CouponID uint64    `gorm:"column:coupon_id"`
	Status   int8      `gorm:"column:status"` // 0-未使用 1-已使用 2-已过期
	OrderID  *uint64   `gorm:"column:order_id"`
	ExpireAt time.Time `gorm:"column:expire_at"`
}

type couponRow struct {
	ID            uint64   `gorm:"column:id"`
	Type          int8     `gorm:"column:type"`          // 1-满减 2-折扣 3-免运费
	DiscountType  int8     `gorm:"column:discount_type"` // 1-固定金额 2-百分比
	DiscountValue float64  `gorm:"column:discount_value"`
	MinAmount     float64  `gorm:"column:min_amount"`
	MaxDiscount   *float64 `gorm:"column:max_discount"`
	Status        int8     `gorm:"column:status"`
}

// lockAndCalcCoupon 在事务内锁定并校验用户券，返回可抵扣金额。
//
// 规则：券必须属于当前用户（token 里的 userID）、未使用、未过期，且订单金额达到门槛；
// 折扣券按百分比计算并受 max_discount 上限约束，最终不超过订单金额本身。
func lockAndCalcCoupon(ctx context.Context, tx *gorm.DB, userID, userCouponID uint64, totalAmount float64) (float64, error) {
	if userCouponID == 0 {
		return 0, nil
	}
	if tx == nil {
		return 0, fmt.Errorf("数据库连接未初始化，无法校验优惠券")
	}

	// FOR UPDATE 行锁：并发下单时同一张券只会被核销一次
	var uc userCouponRow
	if err := tx.WithContext(ctx).Raw(
		"SELECT id, coupon_id, status, order_id, expire_at FROM user_coupon WHERE id = ? AND user_id = ? FOR UPDATE",
		userCouponID, userID,
	).Scan(&uc).Error; err != nil {
		return 0, fmt.Errorf("查询优惠券失败: %w", err)
	}
	if uc.ID == 0 {
		return 0, fmt.Errorf("优惠券不存在")
	}
	if uc.Status != 0 {
		return 0, fmt.Errorf("优惠券已使用或已失效")
	}
	if time.Now().After(uc.ExpireAt) {
		return 0, fmt.Errorf("优惠券已过期")
	}

	var c couponRow
	if err := tx.WithContext(ctx).Raw(
		"SELECT id, type, discount_type, discount_value, min_amount, max_discount, status FROM coupon WHERE id = ?",
		uc.CouponID,
	).Scan(&c).Error; err != nil {
		return 0, fmt.Errorf("查询优惠券规则失败: %w", err)
	}
	if c.ID == 0 || c.Status != 1 {
		return 0, fmt.Errorf("优惠券不可用")
	}
	if totalAmount < c.MinAmount {
		return 0, fmt.Errorf("订单金额未满 %s 元，无法使用该优惠券", strconv.FormatFloat(c.MinAmount, 'f', 2, 64))
	}

	discount := 0.0
	switch c.DiscountType {
	case 1: // 固定金额
		discount = c.DiscountValue
	case 2: // 百分比折扣：discount_value 表示折扣百分比（如 10 = 减 10%）
		discount = totalAmount * c.DiscountValue / 100
		if c.MaxDiscount != nil && discount > *c.MaxDiscount {
			discount = *c.MaxDiscount
		}
	}
	if discount > totalAmount {
		discount = totalAmount
	}
	if discount < 0 {
		discount = 0
	}
	return math.Round(discount*100) / 100, nil
}

// redeemUserCoupon 把用户券标记为已使用并挂到订单上
func redeemUserCoupon(ctx context.Context, tx *gorm.DB, userCouponID, orderID uint64) error {
	res := tx.WithContext(ctx).Exec(
		"UPDATE user_coupon SET status = 1, order_id = ?, used_at = NOW() WHERE id = ? AND status = 0",
		orderID, userCouponID,
	)
	if res.Error != nil {
		return fmt.Errorf("核销优惠券失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("优惠券已被使用")
	}
	return nil
}

// releaseUserCoupon 取消订单时把券退回未使用状态（失败只记日志，不影响取消主流程）
func releaseUserCoupon(ctx context.Context, db *gorm.DB, orderID uint64) {
	if db == nil || orderID == 0 {
		return
	}
	if err := db.WithContext(ctx).Exec(
		"UPDATE user_coupon SET status = 0, order_id = NULL, used_at = NULL WHERE order_id = ? AND status = 1",
		orderID,
	).Error; err != nil {
		logx.Errorf("退还优惠券失败 order_id=%d: %v", orderID, err)
	}
}

// invalidateOrderListCache 失效订单列表缓存。
//
// 列表 key 形如 order:list:{user_id}:{status}:{page}:{page_size}，
// 因此只需清该用户自己的命名空间；管理端「全部订单」（user_id=0）也一并清掉。
// 之前这里用的是 order:list:*，等于每次下单/取消都把所有人的列表缓存清空。
func invalidateOrderListCache(ctx context.Context, cacheOps *cache.CacheOperations, userID uint64) {
	if cacheOps == nil {
		return
	}
	if userID > 0 {
		_ = cacheOps.DeletePattern(ctx, fmt.Sprintf("%s%d:*", cache.KeyPrefixOrderList, userID))
	}
	_ = cacheOps.DeletePattern(ctx, fmt.Sprintf("%s0:*", cache.KeyPrefixOrderList))
}

// invalidateProductCache 扣减库存后失效商品相关缓存（列表用 pattern，详情/SKU 用精确键）
func invalidateProductCache(ctx context.Context, db *gorm.DB, cacheOps *cache.CacheOperations, skuID uint64) {
	if cacheOps == nil {
		return
	}
	// 详情缓存需要 product_id，做一次 SKU 主键查询
	var skuRow struct {
		ProductID uint64 `gorm:"column:product_id"`
	}
	if db != nil {
		if err := db.WithContext(ctx).Table("sku").Select("product_id").Where("id = ?", skuID).Scan(&skuRow).Error; err == nil && skuRow.ProductID > 0 {
			_ = cacheOps.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, skuRow.ProductID))
		}
	}
	_ = cacheOps.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, skuID))
	_ = cacheOps.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
}

// restockStock 回增库存：取消订单时恢复 sku.stock 与 inventory
func restockStock(ctx context.Context, db *gorm.DB, cacheOps *cache.CacheOperations, skuID uint64, quantity int) error {
	if db == nil {
		return nil
	}
	if quantity <= 0 {
		return nil
	}
	// 回增 sku.stock
	res := db.WithContext(ctx).Exec(
		"UPDATE sku SET stock = stock + ? WHERE id = ?",
		quantity, skuID,
	)
	if res.Error != nil {
		return res.Error
	}
	// 同步 inventory：可用库存增加、已售减少（不低于 0）
	_ = db.WithContext(ctx).Exec(
		"UPDATE inventory SET available_stock = available_stock + ?, sold_stock = GREATEST(sold_stock - ?, 0) WHERE sku_id = ?",
		quantity, quantity, skuID,
	).Error
	// 失效商品缓存
	invalidateProductCache(ctx, db, cacheOps, skuID)
	return nil
}
