package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/promotion/model"
	"errors"
	"gorm.io/gorm"
)

// PointsRepository 积分仓库接口
type PointsRepository interface {
	// GetByUserID 获取用户积分
	GetByUserID(ctx context.Context, userID uint64) (*model.Points, error)
	// Create 创建积分记录
	Create(ctx context.Context, points *model.Points) error
	// Update 更新积分
	Update(ctx context.Context, points *model.Points) error
	// AddPoints 增加积分
	AddPoints(ctx context.Context, userID uint64, points int64) error
	// DeductPoints 扣减积分
	DeductPoints(ctx context.Context, userID uint64, points int64) error
}

type pointsRepository struct {
	db *gorm.DB
}

// NewPointsRepository 创建积分仓库
func NewPointsRepository(db *gorm.DB) PointsRepository {
	return &pointsRepository{db: db}
}

// GetByUserID 获取用户积分
func (r *pointsRepository) GetByUserID(ctx context.Context, userID uint64) (*model.Points, error) {
	var points model.Points
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&points).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果不存在，创建新记录
		points = model.Points{
			UserID:    userID,
			Total:     0,
			Used:      0,
			Available: 0,
		}
		err = r.db.WithContext(ctx).Create(&points).Error
		if err != nil {
			return nil, err
		}
		return &points, nil
	}
	if err != nil {
		return nil, err
	}
	return &points, nil
}

// Create 创建积分记录
func (r *pointsRepository) Create(ctx context.Context, points *model.Points) error {
	return r.db.WithContext(ctx).Create(points).Error
}

// Update 更新积分
func (r *pointsRepository) Update(ctx context.Context, points *model.Points) error {
	return r.db.WithContext(ctx).Save(points).Error
}

// AddPoints 增加积分
func (r *pointsRepository) AddPoints(ctx context.Context, userID uint64, points int64) error {
	return r.db.WithContext(ctx).Model(&model.Points{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"total":     gorm.Expr("total + ?", points),
			"available": gorm.Expr("available + ?", points),
		}).Error
}

// DeductPoints 扣减积分
func (r *pointsRepository) DeductPoints(ctx context.Context, userID uint64, points int64) error {
	return r.db.WithContext(ctx).Model(&model.Points{}).
		Where("user_id = ? AND available >= ?", userID, points).
		Updates(map[string]interface{}{
			"used":      gorm.Expr("used + ?", points),
			"available": gorm.Expr("available - ?", points),
		}).Error
}
