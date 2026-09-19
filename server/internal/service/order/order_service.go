package order

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/zhian9/GoForge/server/api/order/v1"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"github.com/zhian9/GoForge/server/internal/service/order/model"
	"github.com/zhian9/GoForge/server/internal/service/order/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderResponse, error) {
	// 用户身份只认 token 里的（由 AuthInterceptor 注入），不信任请求参数里的 user_id
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}

	// 转换请求
	items := make([]service.OrderItemRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.OrderItemRequest{
			SkuID:       uint64(item.SkuId),
			Quantity:    int(item.Quantity),
			ProductName: item.ProductName,
			Price:       parsePrice(item.Price),
		})
	}

	createReq := &service.CreateOrderRequest{
		UserID:          userID,
		AddressID:       uint64(req.AddressId),
		Items:           items,
		OrderType:       int8(req.OrderType),
		Remark:          req.Remark,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		ClearCart:       req.ClearCart,
	}
	if req.CouponId > 0 {
		createReq.CouponID = uint64(req.CouponId)
	}

	// 调用业务逻辑
	resp, err := s.logic.CreateOrder(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	orderProto := convertOrderToProto(resp.Order)
	// 查询订单商品项
	orderItems, err := s.svcCtx.OrderItemRepo.GetByOrderID(ctx, resp.Order.ID)
	if err == nil {
		itemProtos := make([]*v1.OrderItem, 0, len(orderItems))
		for _, item := range orderItems {
			itemProtos = append(itemProtos, convertOrderItemToProto(item))
		}
		orderProto.Items = itemProtos
	} else {
		orderProto.Items = make([]*v1.OrderItem, 0)
	}

	return &v1.CreateOrderResponse{
		Code:    0,
		Message: "创建订单成功",
		Data:    orderProto,
	}, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.GetOrderResponse, error) {
	// 转换请求
	getReq := &service.GetOrderRequest{
		ID:      uint64(req.Id),
		OrderNo: req.OrderNo,
	}

	// 调用业务逻辑
	resp, err := s.logic.GetOrder(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	orderProto := convertOrderToProto(resp.Order)
	itemProtos := make([]*v1.OrderItem, 0, len(resp.OrderItems))
	for _, item := range resp.OrderItems {
		itemProtos = append(itemProtos, convertOrderItemToProto(item))
	}
	orderProto.Items = itemProtos

	return &v1.GetOrderResponse{
		Code:    0,
		Message: "成功",
		Data:    orderProto,
	}, nil
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	// 关键安全修复：只查「当前登录用户」的订单。
	// 修复前这里直接用请求参数里的 user_id，导致不登录、随便传一个 user_id
	// 就能读到他人订单的完整信息（订单号、金额、商品、收货信息）。
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}

	// 转换请求
	listReq := &service.ListOrdersRequest{
		UserID:   userID,
		Status:   int8(req.Status),
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	// 调用业务逻辑
	resp, err := s.logic.ListOrders(ctx, listReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	orders := make([]*v1.Order, 0, len(resp.Orders))
	for _, order := range resp.Orders {
		orders = append(orders, convertOrderToProto(order))
	}

	return &v1.ListOrdersResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.OrderListData{
			List:       orders,
			Page:       int32(resp.Page),
			PageSize:   int32(resp.PageSize),
			Total:      resp.Total,
			TotalPages: int32(resp.TotalPages),
		},
	}, nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.CancelOrderResponse, error) {
	// 转换请求
	cancelReq := &service.CancelOrderRequest{
		ID:      uint64(req.Id),
		OrderNo: req.OrderNo,
		Reason:  req.Reason,
	}

	// 调用业务逻辑
	_, err := s.logic.CancelOrder(ctx, cancelReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.CancelOrderResponse{
		Code:    0,
		Message: "取消订单成功",
	}, nil
}

// ConfirmReceive 确认收货
func (s *OrderService) ConfirmReceive(ctx context.Context, req *v1.ConfirmReceiveRequest) (*v1.ConfirmReceiveResponse, error) {
	// 转换请求
	confirmReq := &service.ConfirmReceiveRequest{
		ID:      uint64(req.Id),
		OrderNo: req.OrderNo,
	}

	// 调用业务逻辑
	_, err := s.logic.ConfirmReceive(ctx, confirmReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ConfirmReceiveResponse{
		Code:    0,
		Message: "确认收货成功",
	}, nil
}

// ShipOrder 发货
func (s *OrderService) ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.ShipOrderResponse, error) {
	shipReq := &service.ShipOrderRequest{
		ID:          uint64(req.Id),
		CompanyCode: req.CompanyCode,
		LogisticsNo: req.LogisticsNo,
	}

	_, err := s.logic.ShipOrder(ctx, shipReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ShipOrderResponse{
		Code:    0,
		Message: "发货成功",
	}, nil
}

// GetStats 获取统计数据
func (s *OrderService) GetStats(ctx context.Context, req *v1.GetStatsRequest) (*v1.GetStatsResponse, error) {
	// 这是「全站经营数据」（总订单数、总销售额、今日订单），不是某个用户自己的统计，
	// 属于管理后台接口。修复前它不需要任何身份就能访问 ——
	// 未登录直接请求 /api/v1/orders/stats 就能拿到全站销售额。
	if err := s.logic.CheckAdmin(ctx); err != nil {
		return nil, convertError(err)
	}

	resp, err := s.logic.GetStats(ctx, &service.GetStatsRequest{})
	if err != nil {
		return nil, convertError(err)
	}
	return &v1.GetStatsResponse{
		Code:        0,
		Message:     "成功",
		TotalOrders: resp.TotalOrders,
		TotalSales:  fmt.Sprintf("%.2f", resp.TotalSales),
		TodayOrders: resp.TodayOrders,
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}

	// 检查是否是 BusinessError
	if bizErr, ok := err.(*apperrors.BusinessError); ok {
		var grpcCode codes.Code
		switch bizErr.Code {
		case apperrors.CodeNotFound, apperrors.CodeOrderNotFound:
			grpcCode = codes.NotFound
		case apperrors.CodeInvalidParam:
			grpcCode = codes.InvalidArgument
		case apperrors.CodeUnauthorized:
			grpcCode = codes.Unauthenticated
		case apperrors.CodeForbidden:
			grpcCode = codes.PermissionDenied
		default:
			grpcCode = codes.Internal
		}
		return status.Error(grpcCode, bizErr.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

// convertOrderToProto 转换订单模型为 Protobuf 消息
func convertOrderToProto(o *model.Order) *v1.Order {
	if o == nil {
		return nil
	}

	// 转换订单项
	items := make([]*v1.OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, convertOrderItemToProto(&item))
	}

	order := &v1.Order{
		Id:              int64(o.ID),
		OrderNo:         o.OrderNo,
		UserId:          int64(o.UserID),
		OrderType:       int32(o.OrderType),
		Status:          int32(o.Status),
		TotalAmount:     formatDecimal(o.TotalAmount),
		PayAmount:       formatDecimal(o.PayAmount),
		DiscountAmount:  formatDecimal(o.DiscountAmount),
		FreightAmount:   formatDecimal(o.FreightAmount),
		ReceiverName:    o.ReceiverName,
		ReceiverPhone:   o.ReceiverPhone,
		ReceiverAddress: o.ReceiverAddress,
		Items:           items, // 添加订单项
		CreatedAt:       formatTime(&o.CreatedAt),
		UpdatedAt:       formatTime(&o.UpdatedAt),
	}

	if o.PaymentMethod != nil {
		order.PaymentMethod = int32(*o.PaymentMethod)
	}
	if o.PaymentTime != nil {
		order.PaymentTime = formatTime(o.PaymentTime)
	}
	if o.DeliveryTime != nil {
		order.DeliveryTime = formatTime(o.DeliveryTime)
	}
	if o.ReceiveTime != nil {
		order.ReceiveTime = formatTime(o.ReceiveTime)
	}
	if o.CancelTime != nil {
		order.CancelTime = formatTime(o.CancelTime)
	}
	if o.CancelReason != nil {
		order.CancelReason = *o.CancelReason
	}
	if o.Remark != nil {
		order.Remark = *o.Remark
	}

	return order
}

// convertOrderItemToProto 转换订单商品项模型为 Protobuf 消息
func convertOrderItemToProto(item *model.OrderItem) *v1.OrderItem {
	if item == nil {
		return nil
	}

	itemProto := &v1.OrderItem{
		Id:          int64(item.ID),
		OrderId:     int64(item.OrderID),
		OrderNo:     item.OrderNo,
		ProductId:   int64(item.ProductID),
		ProductName: item.ProductName,
		SkuId:       int64(item.SkuID),
		SkuCode:     item.SkuCode,
		SkuName:     item.SkuName,
		Price:       formatDecimal(item.Price),
		Quantity:    int32(item.Quantity),
		TotalAmount: formatDecimal(item.TotalAmount),
		CreatedAt:   formatTime(&item.CreatedAt),
	}

	if item.SkuImage != nil {
		itemProto.SkuImage = *item.SkuImage
	}
	if item.SkuSpecs != nil {
		itemProto.SkuSpecs = item.SkuSpecs
	}

	return itemProto
}

// formatDecimal 格式化小数
func formatDecimal(f float64) string {
	return formatFloat(f, 2)
}

// formatFloat 格式化浮点数
func formatFloat(f float64, precision int) string {
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"f", f)
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parsePrice(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
