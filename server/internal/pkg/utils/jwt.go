package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var (
	ErrTokenExpired = errors.New("token已过期")
	ErrTokenInvalid = errors.New("token已无效")
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	// IsAdmin 1 表示管理员。后台需要跨用户查询数据（例如订单列表），
	// 服务端必须能从 token 里判断调用方身份，否则只能退化为「只认 JWT 里的 user_id」，
	// 管理员就永远看不到别人的数据。
	IsAdmin int8 `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID uint64, username, secret string, expire int64) (string, error) {
	return GenerateTokenWithAdmin(userID, username, 0, secret, expire)
}

// GenerateTokenWithAdmin 生成带管理员标识的 JWT Token。
// 旧的 GenerateToken 保留为包装函数，避免影响 auth-check / seckill-check 等调用方。
func GenerateTokenWithAdmin(userID uint64, username string, isAdmin int8, secret string, expire int64) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}
