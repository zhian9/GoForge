package user

import (
	"context"
	"errors"
	"strconv"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/user/v1"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"github.com/zhian9/GoForge/server/internal/service/user/model"
	userservice "github.com/zhian9/GoForge/server/internal/service/user/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	// 转换请求
	registerReq := &userservice.RegisterRequest{
		Username:   req.Username,
		Password:   req.Password,
		Phone:      req.Phone,
		Email:      req.Email,
		VerifyCode: req.VerifyCode,
	}

	// 调用业务逻辑
	resp, err := s.logic.Register(ctx, registerReq)
	if err != nil {
		// 业务错误：返回结构化响应（避免 Gateway 把 gRPC error 转成 HTTP 500 文本）
		if bizErr, ok := err.(*apperrors.BusinessError); ok {
			return &v1.RegisterResponse{
				Code:    int32(bizErr.Code),
				Message: bizErr.Message,
				Data:    nil,
			}, nil
		}
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.RegisterResponse{
		Code:    0,
		Message: "注册成功",
		Data:    convertUserToProto(resp.User),
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// 转换请求
	loginReq := &userservice.LoginRequest{
		Username:   req.Username,
		Password:   req.Password,
		LoginType:  int(req.LoginType),
		VerifyCode: req.VerifyCode,
	}

	// 获取 JWT 配置
	jwtSecret := s.svcCtx.Config.JWT.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-key" // 开发环境默认值
	}
	jwtExpire := s.svcCtx.Config.JWT.Expire
	if jwtExpire == 0 {
		jwtExpire = 7200 // 默认 2 小时
	}

	// 调用业务逻辑
	resp, err := s.logic.Login(ctx, loginReq, jwtSecret, jwtExpire)
	if err != nil {
		// 业务错误：返回结构化响应（避免 Gateway 把 gRPC error 转成 HTTP 500 文本）
		var bizErr *apperrors.BusinessError
		if errors.As(err, &bizErr) {
			return &v1.LoginResponse{
				Code:    int32(bizErr.Code),
				Message: bizErr.Message,
				Data:    nil,
			}, nil
		}
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.LoginResponse{
		Code:    0,
		Message: "登录成功",
		Data: &v1.LoginData{
			User:       convertUserToProto(resp.User),
			Token:      resp.Token,
			ExpireTime: resp.ExpireTime,
			IsAdmin:    int32(resp.User.IsAdmin),
		},
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	// 优先从 context 取 user_id（由 gRPC interceptor 从 Authorization 解析得到）
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		// 兼容：如果请求里带了 user_id，就用请求的
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	// 调用业务逻辑
	getReq := &userservice.GetUserInfoRequest{
		UserID: userID,
	}

	resp, err := s.logic.GetUserInfo(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.GetUserInfoResponse{
		Code:    0,
		Message: "成功",
		Data:    convertUserToProto(resp.User),
	}, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	// 转换请求
	updateReq := &userservice.UpdateUserInfoRequest{
		UserID:   userID,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Birthday: req.Birthday,
		Phone:    req.Phone,
		Email:    req.Email,
	}
	if req.Gender != nil {
		g := int(*req.Gender)
		updateReq.Gender = &g
	}

	// 调用业务逻辑
	resp, err := s.logic.UpdateUserInfo(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.UpdateUserInfoResponse{
		Code:    0,
		Message: "更新成功",
		Data:    convertUserToProto(resp.User),
	}, nil
}

// ListUsers 获取用户列表（管理后台）
func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	// 转换请求
	var status *int8
	if req.Status > 0 {
		statusVal := int8(req.Status)
		status = &statusVal
	}

	listReq := &userservice.ListUsersRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Keyword:  req.Keyword,
		Status:   status,
	}

	// 调用业务逻辑
	resp, err := s.logic.ListUsers(ctx, listReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	users := make([]*v1.User, 0, len(resp.Users))
	for _, user := range resp.Users {
		users = append(users, convertUserToProto(user))
	}

	return &v1.ListUsersResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.ListUsersData{
			Users:    users,
			Total:    int32(resp.Total),
			Page:     int32(resp.Page),
			PageSize: int32(resp.PageSize),
		},
	}, nil
}

// DeleteUser 删除用户（管理后台）
func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	// 转换请求
	deleteReq := &userservice.DeleteUserRequest{
		UserID: uint64(req.Id), // 路径参数 :id 映射到 req.Id
	}

	// 调用业务逻辑
	_, err := s.logic.DeleteUser(ctx, deleteReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.DeleteUserResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// GetAddressList 获取地址列表
func (s *UserService) GetAddressList(ctx context.Context, req *v1.GetAddressListRequest) (*v1.GetAddressListResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	// 调用业务逻辑
	getReq := &userservice.GetAddressListRequest{
		UserID: userID,
	}

	resp, err := s.addressLogic.GetAddressList(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	addresses := make([]*v1.Address, 0, len(resp.Addresses))
	for _, addr := range resp.Addresses {
		addresses = append(addresses, convertAddressToProto(addr))
	}

	return &v1.GetAddressListResponse{
		Code:    0,
		Message: "成功",
		Data:    addresses,
	}, nil
}

// AddAddress 添加地址
func (s *UserService) AddAddress(ctx context.Context, req *v1.AddAddressRequest) (*v1.AddAddressResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	// 转换请求
	addReq := &userservice.AddAddressRequest{
		UserID:        userID,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		PostalCode:    req.PostalCode,
		IsDefault:     int8(req.IsDefault),
	}

	// 调用业务逻辑
	resp, err := s.addressLogic.AddAddress(ctx, addReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.AddAddressResponse{
		Code:    0,
		Message: "添加成功",
		Data:    convertAddressToProto(resp.Address),
	}, nil
}

// UpdateAddress 更新地址
func (s *UserService) UpdateAddress(ctx context.Context, req *v1.UpdateAddressRequest) (*v1.UpdateAddressResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	// 转换请求
	updateReq := &userservice.UpdateAddressRequest{
		ID:            uint64(req.Id),
		UserID:        userID,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		PostalCode:    req.PostalCode,
		IsDefault:     int8(req.IsDefault),
	}

	// 调用业务逻辑
	resp, err := s.addressLogic.UpdateAddress(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.UpdateAddressResponse{
		Code:    0,
		Message: "更新成功",
		Data:    convertAddressToProto(resp.Address),
	}, nil
}

// DeleteAddress 删除地址
func (s *UserService) DeleteAddress(ctx context.Context, req *v1.DeleteAddressRequest) (*v1.DeleteAddressResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	// 转换请求
	deleteReq := &userservice.DeleteAddressRequest{
		ID:     uint64(req.Id),
		UserID: userID,
	}

	// 调用业务逻辑
	_, err := s.addressLogic.DeleteAddress(ctx, deleteReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.DeleteAddressResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// SignIn 每日签到
func (s *UserService) SignIn(ctx context.Context, req *v1.SignInRequest) (*v1.SignInResponse, error) {
	signInReq := &userservice.SignInRequest{
		UserID: uint64(req.UserId),
	}

	resp, err := s.logic.SignIn(ctx, signInReq)
	if err != nil {
		if bizErr, ok := err.(*apperrors.BusinessError); ok {
			return &v1.SignInResponse{
				Code:    int32(bizErr.Code),
				Message: bizErr.Message,
			}, nil
		}
		return nil, convertError(err)
	}

	return &v1.SignInResponse{
		Code:        0,
		Message:     "签到成功",
		AddedPoints: int32(resp.AddedPoints),
		TotalPoints: int32(resp.TotalPoints),
	}, nil
}

// Recharge 余额充值
func (s *UserService) Recharge(ctx context.Context, req *v1.RechargeRequest) (*v1.RechargeResponse, error) {
	amount, _ := strconv.ParseFloat(req.Amount, 64)
	rechargeReq := &userservice.RechargeRequest{
		UserID: uint64(req.UserId),
		Amount: amount,
	}

	resp, err := s.logic.Recharge(ctx, rechargeReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.RechargeResponse{
		Code:    0,
		Message: "充值成功",
		Balance: resp.Balance,
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
		case apperrors.CodeNotFound, apperrors.CodeUserNotFound:
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

// convertUserToProto 转换用户模型为 Protobuf 消息
func convertUserToProto(user *model.User) *v1.User {
	if user == nil {
		return nil
	}

	return &v1.User{
		Id:          int64(user.ID),
		Username:    user.Username,
		Nickname:    user.Nickname,
		Phone:       user.Phone,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Gender:      int32(user.Gender),
		Birthday:    formatDate(user.Birthday),
		Status:      int32(user.Status),
		MemberLevel: int32(user.MemberLevel),
		Points:      int32(user.Points),
		IsAdmin:     int32(user.IsAdmin),
		Balance:     user.Balance,
		CreatedAt:   formatTime(&user.CreatedAt),
		UpdatedAt:   formatTime(&user.UpdatedAt),
	}
}

// convertAddressToProto 转换地址模型为 Protobuf 消息
func convertAddressToProto(addr *model.Address) *v1.Address {
	if addr == nil {
		return nil
	}

	return &v1.Address{
		Id:            int64(addr.ID),
		UserId:        int64(addr.UserID),
		ReceiverName:  addr.ReceiverName,
		ReceiverPhone: addr.ReceiverPhone,
		Province:      addr.Province,
		City:          addr.City,
		District:      addr.District,
		Detail:        addr.Detail,
		PostalCode:    addr.PostalCode,
		IsDefault:     int32(addr.IsDefault),
		CreatedAt:     formatTime(&addr.CreatedAt),
		UpdatedAt:     formatTime(&addr.UpdatedAt),
	}
}

// formatTime 格式化时间为 RFC3339 字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// formatDate 格式化日期为 YYYY-MM-DD 字符串
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
