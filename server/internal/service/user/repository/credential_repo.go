package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/user/model"
)

// CredentialRepository 凭证仓储接口
type CredentialRepository interface {
	Create(ctx context.Context, credential *model.Credential) error
	GetByUserIDAndType(ctx context.Context, userID uint64, credentialType int8) (*model.Credential, error)
	GetByKeyAndType(ctx context.Context, key string, credentialType int8) (*model.Credential, error)
	Update(ctx context.Context, credential *model.Credential) error
	Delete(ctx context.Context, id uint64) error
}

// credentialRepository 凭证仓储实现
type credentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository 创建凭证仓储
func NewCredentialRepository(db *gorm.DB) CredentialRepository {
	return &credentialRepository{
		db: db,
	}
}

// Create 创建凭证
func (r *credentialRepository) Create(ctx context.Context, credential *model.Credential) error {
	return r.db.WithContext(ctx).Create(credential).Error
}

// GetByUserIDAndType 根据用户ID和类型获取凭证
func (r *credentialRepository) GetByUserIDAndType(ctx context.Context, userID uint64, credentialType int8) (*model.Credential, error) {
	var credential model.Credential
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND credential_type = ?", userID, credentialType).
		First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &credential, nil
}

// GetByKeyAndType 根据凭证key和类型获取凭证
func (r *credentialRepository) GetByKeyAndType(ctx context.Context, key string, credentialType int8) (*model.Credential, error) {
	var credential model.Credential
	err := r.db.WithContext(ctx).
		Where("credential_key = ? AND credential_type = ?", key, credentialType).
		First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &credential, nil
}

// Update 更新凭证
func (r *credentialRepository) Update(ctx context.Context, credential *model.Credential) error {
	return r.db.WithContext(ctx).Save(credential).Error
}

// Delete 删除凭证
func (r *credentialRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Credential{}, id).Error
}
