package message

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/message/v1"
	"github.com/zhian9/GoForge/server/internal/service/message/model"
	"github.com/zhian9/GoForge/server/internal/service/message/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MessageService 实现 gRPC 服务接口
type MessageService struct {
	v1.UnimplementedMessageServiceServer
	svcCtx *ServiceContext
	logic  *service.MessageLogic
}

// NewMessageService 创建消息服务
func NewMessageService(svcCtx *ServiceContext) *MessageService {
	logic := service.NewMessageLogic(svcCtx.MessageRepo)

	return &MessageService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(ctx context.Context, req *v1.SendMessageRequest) (*v1.SendMessageResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	sendReq := &service.SendMessageRequest{
		UserID:  userID,
		Type:    int8(req.Type),
		Title:   req.Title,
		Content: req.Content,
		Link:    req.Link,
	}

	err := s.logic.SendMessage(ctx, sendReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.SendMessageResponse{
		Code:    0,
		Message: "发送成功",
	}, nil
}

// GetMessageList 获取消息列表
func (s *MessageService) GetMessageList(ctx context.Context, req *v1.GetMessageListRequest) (*v1.GetMessageListResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	getReq := &service.GetMessageListRequest{
		UserID:   userID,
		Type:     int8(req.Type),
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	resp, err := s.logic.GetMessageList(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	messages := make([]*v1.Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		if m != nil {
			messages = append(messages, convertMessageToProto(m))
		}
	}

	return &v1.GetMessageListResponse{
		Code:    0,
		Message: "成功",
		Data:    messages,
		Total:   int32(resp.Total),
	}, nil
}

// MarkAsRead 标记已读
func (s *MessageService) MarkAsRead(ctx context.Context, req *v1.MarkAsReadRequest) (*v1.MarkAsReadResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	markReq := &service.MarkAsReadRequest{
		UserID:    userID,
		MessageID: uint64(req.MessageId),
	}

	err := s.logic.MarkAsRead(ctx, markReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.MarkAsReadResponse{
		Code:    0,
		Message: "操作成功",
	}, nil
}

// BatchMarkAsRead 批量标记已读
func (s *MessageService) BatchMarkAsRead(ctx context.Context, req *v1.BatchMarkAsReadRequest) (*v1.BatchMarkAsReadResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	messageIDs := make([]uint64, 0, len(req.MessageIds))
	for _, id := range req.MessageIds {
		messageIDs = append(messageIDs, uint64(id))
	}

	batchReq := &service.BatchMarkAsReadRequest{
		UserID:     userID,
		MessageIDs: messageIDs,
	}

	err := s.logic.BatchMarkAsRead(ctx, batchReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.BatchMarkAsReadResponse{
		Code:    0,
		Message: "操作成功",
	}, nil
}

// GetUnreadCount 获取未读数量
func (s *MessageService) GetUnreadCount(ctx context.Context, req *v1.GetUnreadCountRequest) (*v1.GetUnreadCountResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	getReq := &service.GetUnreadCountRequest{
		UserID: userID,
	}

	resp, err := s.logic.GetUnreadCount(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetUnreadCountResponse{
		Code:    0,
		Message: "成功",
		Count:   int32(resp.Count),
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}

// convertMessageToProto 转换消息模型为 Protobuf 消息
func convertMessageToProto(msg *model.Message) *v1.Message {
	if msg == nil {
		return nil
	}

	var link string
	if msg.Link != nil {
		link = *msg.Link
	}

	return &v1.Message{
		Id:        int64(msg.ID),
		UserId:    int64(msg.UserID),
		Type:      int32(msg.Type),
		Title:     msg.Title,
		Content:   msg.Content,
		Link:      link,
		IsRead:    int32(msg.IsRead),
		CreatedAt: formatTime(&msg.CreatedAt),
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
