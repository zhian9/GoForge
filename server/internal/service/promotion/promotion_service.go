package promotion

import (
	"context"
	"encoding/json"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"strconv"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/promotion/v1"
	"github.com/zhian9/GoForge/server/internal/service/promotion/model"
	"github.com/zhian9/GoForge/server/internal/service/promotion/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PromotionService 实现 gRPC 服务接口
type PromotionService struct {
	v1.UnimplementedPromotionServiceServer
	svcCtx *ServiceContext
	logic  *service.PromotionLogic
}

// NewPromotionService 创建营销服务
func NewPromotionService(svcCtx *ServiceContext) *PromotionService {
	logic := service.NewPromotionLogic(
		svcCtx.CouponRepo,
		svcCtx.UserCouponRepo,
		svcCtx.PromotionRepo,
		svcCtx.PointsRepo,
	)

	return &PromotionService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// GetCouponList 获取优惠券列表
func (s *PromotionService) GetCouponList(ctx context.Context, req *v1.GetCouponListRequest) (*v1.GetCouponListResponse, error) {
	getReq := &service.GetCouponListRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	resp, err := s.logic.GetCouponList(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	coupons := make([]*v1.Coupon, 0, len(resp.Coupons))
	for _, coupon := range resp.Coupons {
		coupons = append(coupons, convertCouponToProto(coupon))
	}

	return &v1.GetCouponListResponse{
		Code:    0,
		Message: "成功",
		Data:    coupons,
		Total:   int32(resp.Total),
	}, nil
}

// ReceiveCoupon 领取优惠券
func (s *PromotionService) ReceiveCoupon(ctx context.Context, req *v1.ReceiveCouponRequest) (*v1.ReceiveCouponResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	receiveReq := &service.ReceiveCouponRequest{
		UserID:   userID,
		CouponID: uint64(req.CouponId),
	}

	err := s.logic.ReceiveCoupon(ctx, receiveReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ReceiveCouponResponse{
		Code:    0,
		Message: "领取成功",
	}, nil
}

// GetUserCouponList 获取用户优惠券列表
func (s *PromotionService) GetUserCouponList(ctx context.Context, req *v1.GetUserCouponListRequest) (*v1.GetUserCouponListResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	getReq := &service.GetUserCouponListRequest{
		UserID: userID,
		Status: int8(req.Status),
	}

	resp, err := s.logic.GetUserCouponList(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	userCoupons := make([]*v1.UserCoupon, 0, len(resp.UserCoupons))
	for _, uc := range resp.UserCoupons {
		userCoupons = append(userCoupons, convertUserCouponToProto(uc))
	}

	return &v1.GetUserCouponListResponse{
		Code:    0,
		Message: "成功",
		Data:    userCoupons,
	}, nil
}

// UseCoupon 使用优惠券
func (s *PromotionService) UseCoupon(ctx context.Context, req *v1.UseCouponRequest) (*v1.UseCouponResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	useReq := &service.UseCouponRequest{
		UserID:       userID,
		UserCouponID: uint64(req.UserCouponId),
		OrderID:      uint64(req.OrderId),
	}

	err := s.logic.UseCoupon(ctx, useReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.UseCouponResponse{
		Code:    0,
		Message: "使用成功",
	}, nil
}

// GetPromotionList 获取促销活动列表
func (s *PromotionService) GetPromotionList(ctx context.Context, req *v1.GetPromotionListRequest) (*v1.GetPromotionListResponse, error) {
	getReq := &service.GetPromotionListRequest{
		ProductID:  uint64(req.ProductId),
		CategoryID: uint64(req.CategoryId),
	}

	resp, err := s.logic.GetPromotionList(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	promotions := make([]*v1.Promotion, 0, len(resp.Promotions))
	for _, p := range resp.Promotions {
		promotions = append(promotions, convertPromotionToProto(p))
	}

	return &v1.GetPromotionListResponse{
		Code:    0,
		Message: "成功",
		Data:    promotions,
	}, nil
}

// CalculateDiscount 计算优惠金额
func (s *PromotionService) CalculateDiscount(ctx context.Context, req *v1.CalculateDiscountRequest) (*v1.CalculateDiscountResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	totalAmount, _ := strconv.ParseFloat(req.TotalAmount, 64)

	productIDs := make([]uint64, 0, len(req.ProductIds))
	for _, id := range req.ProductIds {
		productIDs = append(productIDs, uint64(id))
	}

	quantities := make([]int, 0, len(req.Quantities))
	for _, q := range req.Quantities {
		quantities = append(quantities, int(q))
	}

	calcReq := &service.CalculateDiscountRequest{
		UserID:      userID,
		ProductIDs:  productIDs,
		Quantities:  quantities,
		CouponID:    uint64(req.CouponId),
		TotalAmount: totalAmount,
	}

	resp, err := s.logic.CalculateDiscount(ctx, calcReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.CalculateDiscountResponse{
		Code:           0,
		Message:        "成功",
		DiscountAmount: strconv.FormatFloat(resp.DiscountAmount, 'f', 2, 64),
		FinalAmount:    strconv.FormatFloat(resp.FinalAmount, 'f', 2, 64),
	}, nil
}

// GetUserPoints 获取用户积分
func (s *PromotionService) GetUserPoints(ctx context.Context, req *v1.GetUserPointsRequest) (*v1.GetUserPointsResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	getReq := &service.GetUserPointsRequest{
		UserID: userID,
	}

	resp, err := s.logic.GetUserPoints(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetUserPointsResponse{
		Code:    0,
		Message: "成功",
		Points:  resp.Points,
	}, nil
}

// ExchangePoints 积分兑换
func (s *PromotionService) ExchangePoints(ctx context.Context, req *v1.ExchangePointsRequest) (*v1.ExchangePointsResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	exchangeReq := &service.ExchangePointsRequest{
		UserID: userID,
		Points: req.Points,
	}

	err := s.logic.ExchangePoints(ctx, exchangeReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ExchangePointsResponse{
		Code:    0,
		Message: "兑换成功",
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}

// convertCouponToProto 转换优惠券模型为 Protobuf 消息
func convertCouponToProto(coupon *model.Coupon) *v1.Coupon {
	if coupon == nil {
		return nil
	}

	var maxDiscount string
	if coupon.MaxDiscount != nil {
		maxDiscount = strconv.FormatFloat(*coupon.MaxDiscount, 'f', 2, 64)
	}

	return &v1.Coupon{
		Id:             int64(coupon.ID),
		Name:           coupon.Name,
		Type:           int32(coupon.Type),
		DiscountType:   int32(coupon.DiscountType),
		DiscountValue:  strconv.FormatFloat(coupon.DiscountValue, 'f', 2, 64),
		MinAmount:      strconv.FormatFloat(coupon.MinAmount, 'f', 2, 64),
		MaxDiscount:    maxDiscount,
		ValidStartTime: formatTime(&coupon.ValidStartTime),
		ValidEndTime:   formatTime(&coupon.ValidEndTime),
		Status:         int32(coupon.Status),
	}
}

// convertUserCouponToProto 转换用户优惠券模型为 Protobuf 消息
func convertUserCouponToProto(uc *model.UserCoupon) *v1.UserCoupon {
	if uc == nil {
		return nil
	}

	var orderID int64
	if uc.OrderID != nil {
		orderID = int64(*uc.OrderID)
	}

	return &v1.UserCoupon{
		Id:        int64(uc.ID),
		UserId:    int64(uc.UserID),
		CouponId:  int64(uc.CouponID),
		Status:    int32(uc.Status),
		OrderId:   orderID,
		ExpireAt:  formatTime(&uc.ExpireAt),
		CreatedAt: formatTime(&uc.CreatedAt),
	}
}

// convertPromotionToProto 转换促销活动模型为 Protobuf 消息
func convertPromotionToProto(p *model.Promotion) *v1.Promotion {
	if p == nil {
		return nil
	}

	productIDs := make([]int64, 0, len(p.ProductIDs))
	for _, id := range p.ProductIDs {
		productIDs = append(productIDs, int64(id))
	}

	categoryIDs := make([]int64, 0, len(p.CategoryIDs))
	for _, id := range p.CategoryIDs {
		categoryIDs = append(categoryIDs, int64(id))
	}

	ruleJSON, _ := json.Marshal(p.Rule)

	return &v1.Promotion{
		Id:          int64(p.ID),
		Name:        p.Name,
		Type:        int32(p.Type),
		Rule:        string(ruleJSON),
		ProductIds:  productIDs,
		CategoryIds: categoryIDs,
		StartTime:   formatTime(&p.StartTime),
		EndTime:     formatTime(&p.EndTime),
		Status:      int32(p.Status),
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
