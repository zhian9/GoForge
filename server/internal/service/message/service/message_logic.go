package service

import (
	"context"
	"time"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/message/model"
	"github.com/zhian9/GoForge/server/internal/service/message/repository"
)

// MessageLogic 消息业务逻辑
type MessageLogic struct {
	messageRepo repository.MessageRepository
}

// NewMessageLogic 创建消息业务逻辑
func NewMessageLogic(messageRepo repository.MessageRepository) *MessageLogic {
	return &MessageLogic{
		messageRepo: messageRepo,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	UserID  uint64
	Type    int8
	Title   string
	Content string
	Link    string
}

// SendMessage 发送消息
func (l *MessageLogic) SendMessage(ctx context.Context, req *SendMessageRequest) error {
	message := &model.Message{
		UserID:    req.UserID,
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		IsRead:    0,
		CreatedAt: time.Now(),
	}
	if req.Link != "" {
		message.Link = &req.Link
	}

	err := l.messageRepo.Create(ctx, message)
	if err != nil {
		return apperrors.NewInternalError("发送消息失败")
	}

	return nil
}

// GetMessageListRequest 获取消息列表请求
type GetMessageListRequest struct {
	UserID   uint64
	Type     int8
	Page     int
	PageSize int
}

// GetMessageListResponse 获取消息列表响应
type GetMessageListResponse struct {
	Messages []*model.Message
	Total    int64
}

// GetMessageList 获取消息列表
func (l *MessageLogic) GetMessageList(ctx context.Context, req *GetMessageListRequest) (*GetMessageListResponse, error) {
	messages, total, err := l.messageRepo.GetList(ctx, req.UserID, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, apperrors.NewInternalError("获取消息列表失败")
	}

	return &GetMessageListResponse{
		Messages: messages,
		Total:    total,
	}, nil
}

// MarkAsReadRequest 标记已读请求
type MarkAsReadRequest struct {
	UserID    uint64
	MessageID uint64
}

// MarkAsRead 标记已读
func (l *MessageLogic) MarkAsRead(ctx context.Context, req *MarkAsReadRequest) error {
	err := l.messageRepo.MarkAsRead(ctx, req.UserID, req.MessageID)
	if err != nil {
		return apperrors.NewInternalError("标记已读失败")
	}
	return nil
}

// BatchMarkAsReadRequest 批量标记已读请求
type BatchMarkAsReadRequest struct {
	UserID     uint64
	MessageIDs []uint64
}

// BatchMarkAsRead 批量标记已读
func (l *MessageLogic) BatchMarkAsRead(ctx context.Context, req *BatchMarkAsReadRequest) error {
	err := l.messageRepo.BatchMarkAsRead(ctx, req.UserID, req.MessageIDs)
	if err != nil {
		return apperrors.NewInternalError("批量标记已读失败")
	}
	return nil
}

// GetUnreadCountRequest 获取未读数量请求
type GetUnreadCountRequest struct {
	UserID uint64
}

// GetUnreadCountResponse 获取未读数量响应
type GetUnreadCountResponse struct {
	Count int64
}

// GetUnreadCount 获取未读数量
func (l *MessageLogic) GetUnreadCount(ctx context.Context, req *GetUnreadCountRequest) (*GetUnreadCountResponse, error) {
	count, err := l.messageRepo.GetUnreadCount(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("获取未读数量失败")
	}

	return &GetUnreadCountResponse{
		Count: count,
	}, nil
}
