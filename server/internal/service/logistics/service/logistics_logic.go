package service

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/logistics/model"
	"github.com/zhian9/GoForge/server/internal/service/logistics/repository"
)

// LogisticsLogic 物流业务逻辑
type LogisticsLogic struct {
	logisticsRepo repository.LogisticsRepository
}

// NewLogisticsLogic 创建物流业务逻辑
func NewLogisticsLogic(logisticsRepo repository.LogisticsRepository) *LogisticsLogic {
	return &LogisticsLogic{
		logisticsRepo: logisticsRepo,
	}
}

// CreateLogisticsRequest 创建物流单请求
type CreateLogisticsRequest struct {
	OrderID         uint64
	OrderNo         string
	CompanyCode     string
	CompanyName     string
	ReceiverName    string
	ReceiverPhone   string
	ReceiverAddress string
}

// CreateLogisticsResponse 创建物流单响应
type CreateLogisticsResponse struct {
	Logistics *model.Logistics
}

// CreateLogistics 创建物流单
func (l *LogisticsLogic) CreateLogistics(ctx context.Context, req *CreateLogisticsRequest) (*CreateLogisticsResponse, error) {
	// 生成物流单号
	logisticsNo := fmt.Sprintf("L%d%d", time.Now().Unix(), req.OrderID)

	logistics := &model.Logistics{
		OrderID:          req.OrderID,
		OrderNo:          req.OrderNo,
		LogisticsCompany: req.CompanyName,
		LogisticsNo:      logisticsNo,
		ReceiverName:     req.ReceiverName,
		ReceiverPhone:    req.ReceiverPhone,
		ReceiverAddress:  req.ReceiverAddress,
		Status:           0, // 待发货
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := l.logisticsRepo.Create(ctx, logistics)
	if err != nil {
		return nil, apperrors.NewInternalError("创建物流单失败")
	}

	return &CreateLogisticsResponse{
		Logistics: logistics,
	}, nil
}

// GetLogisticsRequest 获取物流信息请求
type GetLogisticsRequest struct {
	OrderID uint64
}

// GetLogisticsResponse 获取物流信息响应
type GetLogisticsResponse struct {
	Logistics *model.Logistics
}

// GetLogistics 获取物流信息
func (l *LogisticsLogic) GetLogistics(ctx context.Context, req *GetLogisticsRequest) (*GetLogisticsResponse, error) {
	logistics, err := l.logisticsRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("物流信息不存在")
	}

	return &GetLogisticsResponse{
		Logistics: logistics,
	}, nil
}

// UpdateLogisticsStatusRequest 更新物流状态请求
type UpdateLogisticsStatusRequest struct {
	LogisticsNo string
	Status      int8
	Remark      string
}

// UpdateLogisticsStatus 更新物流状态
func (l *LogisticsLogic) UpdateLogisticsStatus(ctx context.Context, req *UpdateLogisticsStatusRequest) error {
	logistics, err := l.logisticsRepo.GetByLogisticsNo(ctx, req.LogisticsNo)
	if err != nil {
		return apperrors.NewNotFoundError("物流信息不存在")
	}

	logistics.Status = req.Status
	now := time.Now()
	if req.Status == 1 {
		logistics.ShippedAt = &now
	} else if req.Status == 3 {
		logistics.DeliveredAt = &now
	}

	err = l.logisticsRepo.Update(ctx, logistics)
	if err != nil {
		return apperrors.NewInternalError("更新物流状态失败")
	}

	return nil
}

// QueryTrackingRequest 查询物流轨迹请求
type QueryTrackingRequest struct {
	LogisticsNo string
}

// TrackingNode 物流轨迹节点
type TrackingNode struct {
	Time     string
	Status   string
	Location string
	Remark   string
}

// QueryTrackingResponse 查询物流轨迹响应
type QueryTrackingResponse struct {
	Nodes []*TrackingNode
}

// QueryTracking 查询物流轨迹
func (l *LogisticsLogic) QueryTracking(ctx context.Context, req *QueryTrackingRequest) (*QueryTrackingResponse, error) {
	logistics, err := l.logisticsRepo.GetByLogisticsNo(ctx, req.LogisticsNo)
	if err != nil {
		return nil, apperrors.NewNotFoundError("物流信息不存在")
	}

	// 这里简化处理，实际应该调用第三方物流API获取轨迹
	nodes := []*TrackingNode{
		{
			Time:     logistics.CreatedAt.Format(time.RFC3339),
			Status:   "已创建",
			Location: "",
			Remark:   "物流单已创建",
		},
	}

	if logistics.ShippedAt != nil {
		nodes = append(nodes, &TrackingNode{
			Time:     logistics.ShippedAt.Format(time.RFC3339),
			Status:   "已发货",
			Location: "",
			Remark:   "商品已发出",
		})
	}

	if logistics.DeliveredAt != nil {
		nodes = append(nodes, &TrackingNode{
			Time:     logistics.DeliveredAt.Format(time.RFC3339),
			Status:   "已送达",
			Location: "",
			Remark:   "商品已送达",
		})
	}

	return &QueryTrackingResponse{
		Nodes: nodes,
	}, nil
}

// CalculateFreightRequest 计算运费请求
type CalculateFreightRequest struct {
	Province string
	City     string
	District string
	Weight   float64
	Volume   float64
}

// CalculateFreightResponse 计算运费响应
type CalculateFreightResponse struct {
	Freight float64
}

// CalculateFreight 计算运费
func (l *LogisticsLogic) CalculateFreight(ctx context.Context, req *CalculateFreightRequest) (*CalculateFreightResponse, error) {
	// 简化处理，实际应该根据地区、重量、体积计算
	baseFreight := 10.0
	weightFreight := req.Weight * 2.0
	volumeFreight := req.Volume * 1.5

	freight := baseFreight + weightFreight + volumeFreight
	if freight < 10 {
		freight = 10
	}

	return &CalculateFreightResponse{
		Freight: freight,
	}, nil
}
