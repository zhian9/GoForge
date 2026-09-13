package service

import (
	"context"
	"fmt"
	"time"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/constants"
	"github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"github.com/zhian9/GoForge/server/internal/service/user/model"
	repository "github.com/zhian9/GoForge/server/internal/service/user/repository"
	"gorm.io/gorm"
)

// UserLogic 用户业务逻辑
type UserLogic struct {
	userRepo       repository.UserRepository
	credentialRepo repository.CredentialRepository
	addressRepo    repository.AddressRepository
	cache          *cache.CacheOperations
	db             *gorm.DB
}

// NewUserLogic 创建用户业务逻辑
func NewUserLogic(
	userRepo repository.UserRepository,
	credentialRepo repository.CredentialRepository,
	addressRepo repository.AddressRepository,
	cache *cache.CacheOperations,
	db *gorm.DB,
) *UserLogic {
	return &UserLogic{
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		addressRepo:    addressRepo,
		cache:          cache,
		db:             db,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username   string
	Password   string
	Phone      string
	Email      string
	VerifyCode string
}

// RegisterResponse 注册请求
type RegisterResponse struct {
	UserID   uint64
	Username string
	User     *model.User
}

// Register 用户注册
func (u *UserLogic) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 1 参数验证
	if req.Username == "" {
		return nil, errors.NewInvalidParamError("用户名不能为空")
	}
	if req.Password == "" || len(req.Password) < 6 {
		return nil, errors.NewInvalidParamError("密码长度至少6位")
	}

	// 2.检查用户名是否已存在
	existingUser, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.NewInternalError("查询用户失败:" + err.Error())
	}
	if existingUser != nil {
		return nil, errors.NewError(errors.CodeUserAlreadyExists, "用户名已存在")
	}

	//3.检查手机号是否存在
	if req.Phone != "" {
		existingUser, err = u.userRepo.GetByPhone(ctx, req.Phone)
		if err != nil {
			return nil, errors.NewInternalError("查询用户失败:" + err.Error())
		}
		if existingUser != nil {
			return nil, errors.NewError(errors.CodeUserAlreadyExists, "手机号已被注册")
		}
	}

	//4.检查邮箱是否已经存在
	if req.Email != "" {
		existingUser, err = u.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, errors.NewInternalError("查询用户失败:" + err.Error())
		}
		if existingUser != nil {
			return nil, errors.NewError(errors.CodeUserAlreadyExists, "邮箱已被注册")
		}
	}

	//5. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.NewInternalError("密码加密失败:" + err.Error())
	}

	//6. 创建用户
	now := time.Now()
	user := &model.User{
		Username:    req.Username,
		Status:      constants.UserStatusNormal,
		MemberLevel: constants.MemberLevelNormal,
		Points:      100, // 注册赠送 100 积分
		Balance:     100, // 注册赠送 100 元余额
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	//
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := u.userRepo.CreateWithOmit(ctx, user); err != nil {
		return nil, errors.NewInternalError("创建用户失败:" + err.Error())
	}

	// 7.创建密码凭证
	credential := &model.Credential{
		UserID:          user.ID,
		CredentialType:  1,
		CredentialKey:   req.Username,
		CredentialValue: hashedPassword,
		Extra:           "{}",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := u.credentialRepo.Create(ctx, credential); err != nil {
		_ = u.userRepo.Delete(ctx, user.ID)
		return nil, errors.NewInternalError("创建凭证失败:" + err.Error())
	}

	return &RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
		User:     user,
	}, nil
}

// SignInRequest 签到请求
type SignInRequest struct {
	UserID uint64
}

// SignInResponse 签到响应
type SignInResponse struct {
	AddedPoints int
	TotalPoints int
}

// SignIn 每日签到（+10 积分，每天一次）
func (u *UserLogic) SignIn(ctx context.Context, req *SignInRequest) (*SignInResponse, error) {
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}

	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.NewInternalError("查询用户失败: " + err.Error())
	}
	if user == nil {
		return nil, errors.NewError(errors.CodeUserNotFound, "用户不存在")
	}

	const signInBonus = 10
	if u.cache != nil {
		today := time.Now().Format("20060102")
		signKey := fmt.Sprintf("sign:user:%d:%s", req.UserID, today)
		exists, _ := u.cache.Exists(ctx, signKey)
		if exists {
			return nil, errors.NewError(errors.CodeAlreadyExists, "今日已签到")
		}

		user.Points += signInBonus
		if err := u.userRepo.Update(ctx, user); err != nil {
			return nil, errors.NewInternalError("签到失败: " + err.Error())
		}

		now := time.Now()
		endOfDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		_ = u.cache.Set(ctx, signKey, "1", endOfDay.Sub(now))
		// 积分变化，清理用户信息缓存
		_ = u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID))
	} else {
		user.Points += signInBonus
		if err := u.userRepo.Update(ctx, user); err != nil {
			return nil, errors.NewInternalError("签到失败: " + err.Error())
		}
	}

	return &SignInResponse{AddedPoints: signInBonus, TotalPoints: user.Points}, nil
}

// RechargeRequest 充值请求
type RechargeRequest struct {
	UserID uint64
	Amount float64
}

// RechargeResponse 充值响应
type RechargeResponse struct {
	Balance float64
}

// Recharge 余额充值
func (u *UserLogic) Recharge(ctx context.Context, req *RechargeRequest) (*RechargeResponse, error) {
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}
	if req.Amount <= 0 || req.Amount > 100000 {
		return nil, errors.NewInvalidParamError("充值金额需大于0且不超过100000")
	}
	if u.db == nil {
		return nil, errors.NewInternalError("数据库未初始化")
	}

	// 原子增加余额
	res := u.db.WithContext(ctx).Exec(
		"UPDATE user SET balance = balance + ? WHERE id = ?",
		req.Amount, req.UserID,
	)
	if res.Error != nil {
		return nil, errors.NewInternalError("充值失败")
	}
	if res.RowsAffected == 0 {
		return nil, errors.NewInvalidParamError("用户不存在")
	}

	// 查询充值后余额
	var balance float64
	if err := u.db.WithContext(ctx).Table("user").Select("balance").Where("id = ?", req.UserID).Scan(&balance).Error; err != nil {
		return nil, errors.NewInternalError("查询余额失败")
	}

	// 记余额流水
	_ = u.db.WithContext(ctx).Exec(
		"INSERT INTO balance_log (user_id, order_no, type, amount, before_balance, after_balance, remark, created_at) VALUES (?, '', 1, ?, ?, ?, '充值', ?)",
		req.UserID, req.Amount, balance-req.Amount, balance, time.Now(),
	).Error

	// 清用户缓存
	if u.cache != nil {
		_ = u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID))
	}

	return &RechargeResponse{Balance: balance}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string // 1 用户名 2.手机号 3 邮箱
	Password   string
	LoginType  int    // 1 用户名 2.手机号 3 邮箱
	VerifyCode string //验证码
}

// LoginResponse 登录响应
type LoginResponse struct {
	UserID     uint64
	Username   string
	User       *model.User
	Token      string
	ExpireTime int64
}

// Login 用户登录
func (u *UserLogic) Login(ctx context.Context, req *LoginRequest, jwtSecret string, jwtExpire int64) (*LoginResponse, error) {
	//1.参数验证
	if req.Username == "" {
		return nil, errors.NewInvalidParamError("用户名不能为空")
	}
	if req.Password == "" {
		return nil, errors.NewInvalidParamError("密码不能为空")
	}

	//验证码验证
	if req.VerifyCode != "" {
		//非空检测
	}

	//2.根据登录类型查找用户
	var user *model.User
	var err error

	switch req.LoginType {
	case 1: //用户名登录
		user, err = u.userRepo.GetByUsername(ctx, req.Username)
	case 2: //手机号登录
		user, err = u.userRepo.GetByPhone(ctx, req.Username)
	case 3: //邮箱登录
		user, err = u.userRepo.GetByEmail(ctx, req.Username)
	default:
		return nil, errors.NewInvalidParamError("登录类型错误")
	}

	if err != nil {
		return nil, errors.NewInternalError("查询用户失败:" + err.Error())
	}
	if user == nil {
		return nil, errors.NewError(errors.CodeUserNotFound, "用户不存在")
	}

	//3.检查用户状态 是否封禁
	if user.Status != constants.UserStatusNormal {
		return nil, errors.NewError(errors.CodeForbidden, "用户已被禁用")
	}

	//4.验证密码
	credential, err := u.credentialRepo.GetByUserIDAndType(ctx, user.ID, 1) // 1.密码
	if err != nil {
		return nil, errors.NewInternalError("查询凭证失败:" + err.Error())
	}
	if credential == nil {
		return nil, errors.NewError(errors.CodePasswordError, "密码凭证不存在")
	}

	if !utils.CheckPassword(req.Password, credential.CredentialValue) {
		return nil, errors.NewError(errors.CodePasswordError, "密码错误")
	}

	//5.生产jwt Token
	token, err := utils.GenerateToken(user.ID, user.Username, jwtSecret, jwtExpire)
	if err != nil {
		return nil, errors.NewInternalError("生产Token失败:" + err.Error())
	}

	expireTime := time.Now().Add(time.Duration(jwtExpire) * time.Second).Unix()

	//缓存会话信息
	if u.cache != nil {
		sessionKey := cache.BuildKey(cache.KeyPrefixUserSession, token)
		sessionData := map[string]interface{}{
			"user_id":   user.ID,
			"username":  user.Username,
			"expire_at": expireTime,
		}
		_ = u.cache.Set(ctx, sessionKey, sessionData, time.Duration(jwtExpire)*time.Second)

		//缓存用户信息
		userKey := cache.BuildKey(cache.KeyPrefixUserInfo, user.ID)
		_ = u.cache.Set(ctx, userKey, user, 30*time.Minute)
	}

	return &LoginResponse{
		UserID:     user.ID,
		Username:   user.Username,
		User:       user,
		Token:      token,
		ExpireTime: expireTime,
	}, nil
}

// GetUserInfoRequest 获取用户信息请求
type GetUserInfoRequest struct {
	UserID uint64
}

// GetUserInfoResponse 获取用户信息响应
type GetUserInfoResponse struct {
	User *model.User
}

// GetUserInfo 获取用户信息
func (u *UserLogic) GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	//1.参数验证
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}

	//2.尝试从缓存获取
	if u.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID)
		var user model.User
		if err := u.cache.GetJSON(ctx, cacheKey, &user); err == nil {
			return &GetUserInfoResponse{User: &user}, nil
		}
	}

	//3.从数据库查询
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.NewInternalError("查询用户失败: " + err.Error())
	}
	if user == nil {
		return nil, errors.NewError(errors.CodeUserNotFound, "用户不存在")
	}

	//写入缓存
	if u.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID)
		_ = u.cache.Set(ctx, cacheKey, user, 30*time.Minute)
	}

	return &GetUserInfoResponse{User: user}, nil
}

// UpdateUserInfoRequest 更新用户信息请求
type UpdateUserInfoRequest struct {
	UserID   uint64
	Nickname string
	Avatar   string
	Gender   *int
	Birthday string
	Phone    string
	Email    string
}

// UpdateUserInfoResponse 更新用户信息响应
type UpdateUserInfoResponse struct {
	User *model.User
}

// UpdateUserInfo 更新用户信息
func (u *UserLogic) UpdateUserInfo(ctx context.Context, req *UpdateUserInfoRequest) (*UpdateUserInfoResponse, error) {
	// 1.验证参数
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}

	// 2. 获取用户
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.NewInternalError("用户查询失败: " + err.Error())
	}
	if user == nil {
		return nil, errors.NewError(errors.CodeUserNotFound, "用户不存在")
	}

	//3. 更新用户信息
	updated := false
	if req.Nickname != "" {
		user.Nickname = req.Nickname
		updated = true
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
		updated = true
	}
	if req.Gender != nil {
		user.Gender = int8(*req.Gender)
		updated = true
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return nil, errors.NewInvalidParamError("生日格式错误，应为 YYYY-MM-DD")
		}
		user.Birthday = &birthday
		updated = true
	}
		if req.Phone != "" && req.Phone != user.Phone {
			existing, err := u.userRepo.GetByPhone(ctx, req.Phone)
			if err != nil {
				return nil, errors.NewInternalError("校验手机号失败: " + err.Error())
			}
			if existing != nil && existing.ID != user.ID {
				return nil, errors.NewError(errors.CodeUserAlreadyExists, "手机号已被其他用户使用")
			}
			user.Phone = req.Phone
			updated = true
		}
		if req.Email != "" && req.Email != user.Email {
			existing, err := u.userRepo.GetByEmail(ctx, req.Email)
			if err != nil {
				return nil, errors.NewInternalError("校验邮箱失败: " + err.Error())
			}
			if existing != nil && existing.ID != user.ID {
				return nil, errors.NewError(errors.CodeUserAlreadyExists, "邮箱已被其他用户使用")
			}
			user.Email = req.Email
			updated = true
		}

	if !updated {
		return &UpdateUserInfoResponse{User: user}, nil
	}

	//4.保存更新
	user.UpdatedAt = time.Now()
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, errors.NewInternalError("更新用户信息失败: " + err.Error())
	}

	// 删除缓存（确保下次 GetUserInfo 能拉到最新数据）
	if u.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID)
		if err := u.cache.Delete(ctx, cacheKey); err != nil {
			fmt.Printf("[UpdateUserInfo] 删除用户缓存失败: %v\n", err)
		}
		// 也删除用户地址缓存，避免不一致
		addrCacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = u.cache.Delete(ctx, addrCacheKey)
	}

	return &UpdateUserInfoResponse{
		User: user,
	}, nil
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int
	PageSize int
	Keyword  string
	Status   *int8
}

// ListUsersResponse 用户列表响应

type ListUsersResponse struct {
	Users    []*model.User
	Total    int64
	Page     int
	PageSize int
}

// ListUsers 获取用户列表(管理后台)
func (u *UserLogic) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	//0.数据库验证
	if u.userRepo == nil {
		return nil, errors.NewInternalError("数据库连接未初始化")
	}

	//1.参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	//2.查询
	users, total, err := u.userRepo.List(ctx, req.Page, req.PageSize, req.Keyword, req.Status)
	if err != nil {
		return nil, errors.NewInternalError("查询用户列表失败: " + err.Error())
	}

	return &ListUsersResponse{
		Users:    users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	UserID uint64
}

// DeleteUserResponse 删除用户响应
type DeleteUserResponse struct {
}

// DeleteUser 删除用户 (管理后台)
func (u *UserLogic) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	//0.数据库验证
	if u.userRepo == nil {
		return nil, errors.NewInternalError("数据库连接未初始化")
	}

	// 1.参数验证
	if req.UserID == 0 {
		return nil, errors.NewInvalidParamError("用户ID不能为空")
	}

	//检查用户是否存在
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.NewInternalError("查询用户失败: " + err.Error())
	}
	if user == nil {
		return nil, errors.NewError(errors.CodeUserNotFound, "用户不存在")
	}

	// 删除用户 (软删除)
	if err := u.userRepo.Delete(ctx, req.UserID); err != nil {
		return nil, errors.NewInternalError("删除用户失败: " + err.Error())
	}

	return &DeleteUserResponse{}, nil
}
