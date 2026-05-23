package service

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/user/model"
	"github.com/zhian9/GoForge/server/internal/service/user/repository"
	"time"
)

// AddressLogic 地址业务逻辑
type AddressLogic struct {
	addressRepo repository.AddressRepository
	cache       *cache.CacheOperations
}

// NewAddressLogic 创建地址业务逻辑
func NewAddressLogic(addressRepo repository.AddressRepository, cache *cache.CacheOperations) *AddressLogic {
	return &AddressLogic{
		addressRepo: addressRepo,
		cache:       cache,
	}
}

// GetAddressListRequest 获取地址列表请求
type GetAddressListRequest struct {
	UserID uint64
}

// GetAddressListResponse 获取地址列表请求
type GetAddressListResponse struct {
	Addresses []*model.Address
}

// GetAddressList 获取地址列表
func (a *AddressLogic) GetAddressList(ctx context.Context, req *GetAddressListRequest) (*GetAddressListResponse, error) {
	//1.参数验证
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}

	//2.尝试从缓存获取
	if a.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		var addresses []*model.Address
		if err := a.cache.GetJSON(ctx, cacheKey, &addresses); err == nil {
			return &GetAddressListResponse{Addresses: addresses}, nil
		}
	}

	//3.从数据库查询
	addresses, err := a.addressRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, errors.NewInternalError("查询地址列表失败: " + err.Error())
	}

	//写入缓存
	if a.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = a.cache.Set(ctx, cacheKey, addresses, 1*time.Hour)
	}

	return &GetAddressListResponse{Addresses: addresses}, nil
}

// AddAddressRequest 添加地址请求
type AddAddressRequest struct {
	UserID        uint64
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	Detail        string
	PostalCode    string
	IsDefault     int8
}

// AddAddressResponse 添加地址响应
type AddAddressResponse struct {
	Address *model.Address
}

// AddAddress 添加地址
func (a *AddressLogic) AddAddress(ctx context.Context, req *AddAddressRequest) (*AddAddressResponse, error) {
	//1.参数验证
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}
	if req.ReceiverName == "" {
		return nil, errors.NewInvalidParamError("收货人姓名不能为空")
	}
	if req.ReceiverPhone == "" {
		return nil, errors.NewInvalidParamError("收货人电话不能为空")
	}
	if req.Province == "" || req.City == "" || req.District == "" {
		return nil, errors.NewInvalidParamError("地址信息不完整")
	}
	if req.Detail == "" {
		return nil, errors.NewInvalidParamError("详细地址不能为空")
	}

	//2.如果设置为默认地址， 先取消其他默认地址
	if req.IsDefault == 1 {
		if err := a.addressRepo.SetDefault(ctx, req.UserID, 0); err != nil {
			//若设置失败，继续创建地址，但不设为默认
			req.IsDefault = 0
		}
	}

	//3.添加
	address := &model.Address{
		UserID:        req.UserID,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
	}

	if err := a.addressRepo.Create(ctx, address); err != nil {
		return nil, errors.NewInternalError("创建地址失败: " + err.Error())
	}

	//删除缓存
	if a.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = a.cache.Delete(ctx, cacheKey)
	}

	return &AddAddressResponse{Address: address}, nil
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	ID            uint64
	UserID        uint64
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	Detail        string
	PostalCode    string
	IsDefault     int8
}

// UpdateAddressResponse 更新地址响应
type UpdateAddressResponse struct {
	Address *model.Address
}

// UpdateAddress 更新地址
func (a *AddressLogic) UpdateAddress(ctx context.Context, req *UpdateAddressRequest) (*UpdateAddressResponse, error) {
	//1.参数验证
	if req.ID == 0 {
		return nil, errors.NewInvalidParamError("地址ID不能为空")
	}
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}

	//2.获取原地址
	address, err := a.addressRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errors.NewInternalError("查询地址失败: " + err.Error())
	}
	if address == nil {
		return nil, errors.NewError(errors.CodeNotFound, "地址不存在")
	}

	//3.验证地址属于该用户
	if address.UserID != req.UserID {
		return nil, errors.NewError(errors.CodeForbidden, "无权操作此地址/该地址不属于当前用户")
	}

	//4.更新地址信息
	if req.ReceiverName != "" {
		address.ReceiverName = req.ReceiverName
	}
	if req.ReceiverPhone != "" {
		address.ReceiverPhone = req.ReceiverPhone
	}
	if req.Province != "" {
		address.Province = req.Province
	}
	if req.City != "" {
		address.City = req.City
	}
	if req.District != "" {
		address.District = req.District
	}
	if req.Detail != "" {
		address.Detail = req.Detail
	}
	if req.PostalCode != "" {
		address.PostalCode = req.PostalCode
	}

	//5. 如果设置为默认地址
	if req.IsDefault == 1 && address.IsDefault != 1 {
		if err := a.addressRepo.SetDefault(ctx, req.UserID, req.ID); err != nil {
			return nil, errors.NewInternalError("设置默认地址失败: " + err.Error())
		}
		address.IsDefault = 1
	}

	if err := a.addressRepo.Update(ctx, address); err != nil {
		return nil, errors.NewInternalError("更新地址失败: " + err.Error())
	}

	//6.删除缓存
	if a.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = a.cache.Delete(ctx, cacheKey)
	}

	return &UpdateAddressResponse{Address: address}, nil
}

// DeleteAddressRequest 删除地址请求
type DeleteAddressRequest struct {
	ID     uint64
	UserID uint64
}

// DeleteAddressResponse 删除地址响应
type DeleteAddressResponse struct {
}

// DeleteAddress 删除地址
func (a *AddressLogic) DeleteAddress(ctx context.Context, req *DeleteAddressRequest) (*DeleteAddressResponse, error) {
	//1.参数验证
	if req.ID == 0 {
		return nil, errors.NewInvalidParamError("地址ID不能为空")
	}
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}

	//2.获取原地址
	address, err := a.addressRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errors.NewInternalError("查询地址失败: " + err.Error())
	}
	if address == nil {
		return nil, errors.NewError(errors.CodeNotFound, "地址不存在")
	}

	//3.验证地址属于该用户
	if address.UserID != req.UserID {
		return nil, errors.NewError(errors.CodeForbidden, "无权操作此地址")
	}

	//4.删除地址
	if err := a.addressRepo.Delete(ctx, req.ID); err != nil {
		return nil, errors.NewInternalError("删除地址失败: " + err.Error())
	}

	//5.删除缓存
	if a.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = a.cache.Delete(ctx, cacheKey)
	}

	return &DeleteAddressResponse{}, nil
}
