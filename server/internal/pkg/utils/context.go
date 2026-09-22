package utils

import (
	"context"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
	isAdminKey  contextKey = "is_admin"
)

// WithUserID 将用户ID添加到Context
func WithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID 从Context获取用户ID
func GetUserID(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(userIDKey).(uint64)
	return userID, ok
}

// WithUsername 将用户名添加到Context
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

// GetUsername 从Context获取用户名
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}

// WithIsAdmin 将管理员标识添加到Context
func WithIsAdmin(ctx context.Context, isAdmin int8) context.Context {
	return context.WithValue(ctx, isAdminKey, isAdmin)
}

// GetIsAdmin 从Context获取管理员标识（1 表示管理员）
func GetIsAdmin(ctx context.Context) (int8, bool) {
	isAdmin, ok := ctx.Value(isAdminKey).(int8)
	return isAdmin, ok
}
