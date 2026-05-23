package service

import (
	"context"
	"time"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/promotion/model"
	"github.com/zhian9/GoForge/server/internal/service/promotion/repository"
)

// PromotionLogic 营销业务逻辑
type PromotionLogic struct {
	couponRepo     repository.CouponRepository
	userCouponRepo repository.UserCouponRepository
	promotionRepo  repository.PromotionRepository
	pointsRepo     repository.PointsRepository
}

// NewPromotionLogic 创建营销业务逻辑
func NewPromotionLogic(
	couponRepo repository.CouponRepository,
	userCouponRepo repository.UserCouponRepository,
	promotionRepo repository.PromotionRepository,
	pointsRepo repository.PointsRepository,
) *PromotionLogic {
	return &PromotionLogic{
		couponRepo:     couponRepo,
		userCouponRepo: userCouponRepo,
		promotionRepo:  promotionRepo,
		pointsRepo:     pointsRepo,
	}
}

// GetCouponListRequest 获取优惠券列表请求
type GetCouponListRequest struct {
	Page     int
	PageSize int
}

// GetCouponListResponse 获取优惠券列表响应
type GetCouponListResponse struct {
	Coupons []*model.Coupon
	Total   int64
}

// GetCouponList 获取优惠券列表
func (l *PromotionLogic) GetCouponList(ctx context.Context, req *GetCouponListRequest) (*GetCouponListResponse, error) {
	coupons, total, err := l.couponRepo.GetList(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, apperrors.NewInternalError("获取优惠券列表失败")
	}

	return &GetCouponListResponse{
		Coupons: coupons,
		Total:   total,
	}, nil
}

// ReceiveCouponRequest 领取优惠券请求
type ReceiveCouponRequest struct {
	UserID   uint64
	CouponID uint64
}

// ReceiveCoupon 领取优惠券
func (l *PromotionLogic) ReceiveCoupon(ctx context.Context, req *ReceiveCouponRequest) error {
	// 获取优惠券信息
	coupon, err := l.couponRepo.GetByID(ctx, req.CouponID)
	if err != nil {
		return apperrors.NewNotFoundError("优惠券不存在")
	}

	// 检查优惠券状态
	if coupon.Status != 1 {
		return apperrors.NewError(7001, "优惠券已禁用")
	}

	// 检查是否在有效期内
	now := time.Now()
	if now.Before(coupon.ValidStartTime) || now.After(coupon.ValidEndTime) {
		return apperrors.NewError(7002, "优惠券不在有效期内")
	}

	// 检查是否还有剩余
	if coupon.TotalCount > 0 && coupon.UsedCount >= coupon.TotalCount {
		return apperrors.NewError(7003, "优惠券已领完")
	}

	// 检查用户是否已达到限领数量
	count, err := l.userCouponRepo.CountByUserAndCoupon(ctx, req.UserID, req.CouponID)
	if err != nil {
		return apperrors.NewInternalError("检查领取数量失败")
	}
	if int64(coupon.PerUserLimit) > 0 && count >= int64(coupon.PerUserLimit) {
		return apperrors.NewError(7004, "已达到限领数量")
	}

	// 创建用户优惠券
	userCoupon := &model.UserCoupon{
		UserID:    req.UserID,
		CouponID:  req.CouponID,
		Status:    0, // 未使用
		ExpireAt:  coupon.ValidEndTime,
		CreatedAt: time.Now(),
	}

	err = l.userCouponRepo.Create(ctx, userCoupon)
	if err != nil {
		return apperrors.NewInternalError("领取优惠券失败")
	}

	// 更新优惠券已使用数量
	coupon.UsedCount++
	_ = l.couponRepo.Update(ctx, coupon)

	return nil
}

// GetUserCouponListRequest 获取用户优惠券列表请求
type GetUserCouponListRequest struct {
	UserID uint64
	Status int8
}

// GetUserCouponListResponse 获取用户优惠券列表响应
type GetUserCouponListResponse struct {
	UserCoupons []*model.UserCoupon
}

// GetUserCouponList 获取用户优惠券列表
func (l *PromotionLogic) GetUserCouponList(ctx context.Context, req *GetUserCouponListRequest) (*GetUserCouponListResponse, error) {
	userCoupons, err := l.userCouponRepo.GetByUserID(ctx, req.UserID, req.Status)
	if err != nil {
		return nil, apperrors.NewInternalError("获取用户优惠券列表失败")
	}

	return &GetUserCouponListResponse{
		UserCoupons: userCoupons,
	}, nil
}

// UseCouponRequest 使用优惠券请求
type UseCouponRequest struct {
	UserID       uint64
	UserCouponID uint64
	OrderID      uint64
}

// UseCoupon 使用优惠券
func (l *PromotionLogic) UseCoupon(ctx context.Context, req *UseCouponRequest) error {
	// 获取用户优惠券
	userCoupons, err := l.userCouponRepo.GetByUserID(ctx, req.UserID, 0)
	if err != nil {
		return apperrors.NewInternalError("获取用户优惠券失败")
	}

	var userCoupon *model.UserCoupon
	for _, uc := range userCoupons {
		if uc.ID == req.UserCouponID {
			userCoupon = uc
			break
		}
	}

	if userCoupon == nil {
		return apperrors.NewNotFoundError("用户优惠券不存在")
	}

	if userCoupon.Status != 0 {
		return apperrors.NewError(7005, "优惠券已使用或已过期")
	}

	// 检查是否过期
	if time.Now().After(userCoupon.ExpireAt) {
		userCoupon.Status = 2 // 已过期
		_ = l.userCouponRepo.Update(ctx, userCoupon)
		return apperrors.NewError(7006, "优惠券已过期")
	}

	// 更新优惠券状态
	now := time.Now()
	userCoupon.Status = 1 // 已使用
	userCoupon.OrderID = &req.OrderID
	userCoupon.UsedAt = &now

	err = l.userCouponRepo.Update(ctx, userCoupon)
	if err != nil {
		return apperrors.NewInternalError("使用优惠券失败")
	}

	return nil
}

// GetPromotionListRequest 获取促销活动列表请求
type GetPromotionListRequest struct {
	ProductID  uint64
	CategoryID uint64
}

// GetPromotionListResponse 获取促销活动列表响应
type GetPromotionListResponse struct {
	Promotions []*model.Promotion
}

// GetPromotionList 获取促销活动列表
func (l *PromotionLogic) GetPromotionList(ctx context.Context, req *GetPromotionListRequest) (*GetPromotionListResponse, error) {
	promotions, err := l.promotionRepo.GetList(ctx, req.ProductID, req.CategoryID)
	if err != nil {
		return nil, apperrors.NewInternalError("获取促销活动列表失败")
	}

	return &GetPromotionListResponse{
		Promotions: promotions,
	}, nil
}

// CalculateDiscountRequest 计算优惠金额请求
type CalculateDiscountRequest struct {
	UserID      uint64
	ProductIDs  []uint64
	Quantities  []int
	CouponID    uint64
	TotalAmount float64
}

// CalculateDiscountResponse 计算优惠金额响应
type CalculateDiscountResponse struct {
	DiscountAmount float64
	FinalAmount    float64
}

// CalculateDiscount 计算优惠金额
func (l *PromotionLogic) CalculateDiscount(ctx context.Context, req *CalculateDiscountRequest) (*CalculateDiscountResponse, error) {
	discountAmount := 0.0

	// 如果使用了优惠券
	if req.CouponID > 0 {
		coupon, err := l.couponRepo.GetByID(ctx, req.CouponID)
		if err == nil {
			// 检查是否满足最低使用金额
			if req.TotalAmount >= coupon.MinAmount {
				if coupon.DiscountType == 1 {
					// 固定金额
					discountAmount = coupon.DiscountValue
				} else if coupon.DiscountType == 2 {
					// 百分比折扣
					discountAmount = req.TotalAmount * coupon.DiscountValue / 100
					if coupon.MaxDiscount != nil && discountAmount > *coupon.MaxDiscount {
						discountAmount = *coupon.MaxDiscount
					}
				}
			}
		}
	}

	// 计算最终金额
	finalAmount := req.TotalAmount - discountAmount
	if finalAmount < 0 {
		finalAmount = 0
	}

	return &CalculateDiscountResponse{
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
	}, nil
}

// GetUserPointsRequest 获取用户积分请求
type GetUserPointsRequest struct {
	UserID uint64
}

// GetUserPointsResponse 获取用户积分响应
type GetUserPointsResponse struct {
	Points int64
}

// GetUserPoints 获取用户积分
func (l *PromotionLogic) GetUserPoints(ctx context.Context, req *GetUserPointsRequest) (*GetUserPointsResponse, error) {
	points, err := l.pointsRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("获取用户积分失败")
	}

	return &GetUserPointsResponse{
		Points: points.Available,
	}, nil
}

// ExchangePointsRequest 积分兑换请求
type ExchangePointsRequest struct {
	UserID uint64
	Points int64
}

// ExchangePoints 积分兑换
func (l *PromotionLogic) ExchangePoints(ctx context.Context, req *ExchangePointsRequest) error {
	if req.Points <= 0 {
		return apperrors.NewInvalidParamError("积分数量必须大于0")
	}

	// 检查用户积分是否充足
	points, err := l.pointsRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return apperrors.NewInternalError("获取用户积分失败")
	}

	if points.Available < req.Points {
		return apperrors.NewError(7007, "积分不足")
	}

	// 扣减积分
	err = l.pointsRepo.DeductPoints(ctx, req.UserID, req.Points)
	if err != nil {
		return apperrors.NewInternalError("积分兑换失败")
	}

	return nil
}
