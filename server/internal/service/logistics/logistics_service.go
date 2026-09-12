package logistics

import (
	"context"
	"strconv"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/logistics/v1"
	"github.com/zhian9/GoForge/server/internal/service/logistics/model"
	"github.com/zhian9/GoForge/server/internal/service/logistics/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LogisticsService 实现 gRPC 服务接口
type LogisticsService struct {
	v1.UnimplementedLogisticsServiceServer
	svcCtx *ServiceContext
	logic  *service.LogisticsLogic
}

// NewLogisticsService 创建物流服务
func NewLogisticsService(svcCtx *ServiceContext) *LogisticsService {
	logic := service.NewLogisticsLogic(svcCtx.LogisticsRepo)

	return &LogisticsService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// CreateLogistics 创建物流单
func (s *LogisticsService) CreateLogistics(ctx context.Context, req *v1.CreateLogisticsRequest) (*v1.CreateLogisticsResponse, error) {
	createReq := &service.CreateLogisticsRequest{
		OrderID:         uint64(req.OrderId),
		OrderNo:         req.OrderNo,
		CompanyCode:     req.CompanyCode,
		CompanyName:     req.CompanyCode, // 简化处理
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
	}

	resp, err := s.logic.CreateLogistics(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.CreateLogisticsResponse{
		Code:    0,
		Message: "创建成功",
		Data:    convertLogisticsToProto(resp.Logistics),
	}, nil
}

// GetLogistics 获取物流信息
func (s *LogisticsService) GetLogistics(ctx context.Context, req *v1.GetLogisticsRequest) (*v1.GetLogisticsResponse, error) {
	getReq := &service.GetLogisticsRequest{
		OrderID: uint64(req.OrderId),
	}

	resp, err := s.logic.GetLogistics(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetLogisticsResponse{
		Code:    0,
		Message: "成功",
		Data:    convertLogisticsToProto(resp.Logistics),
	}, nil
}

// UpdateLogisticsStatus 更新物流状态
func (s *LogisticsService) UpdateLogisticsStatus(ctx context.Context, req *v1.UpdateLogisticsStatusRequest) (*v1.UpdateLogisticsStatusResponse, error) {
	updateReq := &service.UpdateLogisticsStatusRequest{
		LogisticsNo: req.LogisticsNo,
		Status:      int8(req.Status),
		Remark:      req.Remark,
	}

	err := s.logic.UpdateLogisticsStatus(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.UpdateLogisticsStatusResponse{
		Code:    0,
		Message: "更新成功",
	}, nil
}

// QueryTracking 查询物流轨迹
func (s *LogisticsService) QueryTracking(ctx context.Context, req *v1.QueryTrackingRequest) (*v1.QueryTrackingResponse, error) {
	queryReq := &service.QueryTrackingRequest{
		LogisticsNo: req.LogisticsNo,
	}

	resp, err := s.logic.QueryTracking(ctx, queryReq)
	if err != nil {
		return nil, convertError(err)
	}

	nodes := make([]*v1.TrackingNode, 0, len(resp.Nodes))
	for _, node := range resp.Nodes {
		nodes = append(nodes, &v1.TrackingNode{
			Time:     node.Time,
			Status:   node.Status,
			Location: node.Location,
			Remark:   node.Remark,
		})
	}

	return &v1.QueryTrackingResponse{
		Code:    0,
		Message: "成功",
		Data:    nodes,
	}, nil
}

// CalculateFreight 计算运费
func (s *LogisticsService) CalculateFreight(ctx context.Context, req *v1.CalculateFreightRequest) (*v1.CalculateFreightResponse, error) {
	calcReq := &service.CalculateFreightRequest{
		Province: req.Province,
		City:     req.City,
		District: req.District,
		Weight:   req.Weight,
		Volume:   req.Volume,
	}

	resp, err := s.logic.CalculateFreight(ctx, calcReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.CalculateFreightResponse{
		Code:    0,
		Message: "成功",
		Freight: strconv.FormatFloat(resp.Freight, 'f', 2, 64),
	}, nil
}

// ListLogistics 获取物流列表
func (s *LogisticsService) ListLogistics(ctx context.Context, req *v1.ListLogisticsRequest) (*v1.ListLogisticsResponse, error) {
	listReq := &service.ListLogisticsRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		OrderNo:  req.OrderNo,
	}

	resp, err := s.logic.ListLogistics(ctx, listReq)
	if err != nil {
		return nil, convertError(err)
	}

	list := make([]*v1.Logistics, 0, len(resp.Logistics))
	for _, lg := range resp.Logistics {
		list = append(list, convertLogisticsToProto(lg))
	}

	return &v1.ListLogisticsResponse{
		Code:    0,
		Message: "成功",
		Data:    list,
		Total:   resp.Total,
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}

// convertLogisticsToProto 转换物流模型为 Protobuf 消息
func convertLogisticsToProto(l *model.Logistics) *v1.Logistics {
	if l == nil {
		return nil
	}

	var senderName, senderPhone, senderAddress string
	if l.SenderName != nil {
		senderName = *l.SenderName
	}
	if l.SenderPhone != nil {
		senderPhone = *l.SenderPhone
	}
	if l.SenderAddress != nil {
		senderAddress = *l.SenderAddress
	}
	var currentLocation string
	if l.CurrentLocation != nil {
		currentLocation = *l.CurrentLocation
	}

	return &v1.Logistics{
		Id:              int64(l.ID),
		OrderId:         int64(l.OrderID),
		OrderNo:         l.OrderNo,
		LogisticsNo:     l.LogisticsNo,
		CompanyCode:     l.LogisticsCompany,
		CompanyName:     l.LogisticsCompany,
		Status:          int32(l.Status),
		ReceiverName:    l.ReceiverName,
		ReceiverPhone:   l.ReceiverPhone,
		ReceiverAddress: l.ReceiverAddress,
		SenderName:      senderName,
		SenderPhone:     senderPhone,
		SenderAddress:   senderAddress,
		CreatedAt:       formatTime(&l.CreatedAt),
		UpdatedAt:       formatTime(&l.UpdatedAt),
		CurrentLocation: currentLocation,
		ShippedAt:       formatTime(l.ShippedAt),
		DeliveredAt:     formatTime(l.DeliveredAt),
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
