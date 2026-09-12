package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/logistics/model"
	"github.com/zhian9/GoForge/server/internal/service/logistics/repository"
)

// 物流公司编码 -> 名称映射
var companyNameMap = map[string]string{
	"SF":  "顺丰速运",
	"YTO": "圆通速递",
	"ZTO": "中通快递",
	"STO": "申通快递",
	"YD":  "韵达快递",
	"JD":  "京东物流",
	"YZ":  "邮政EMS",
}

// 物流状态 -> 文案
var statusTextMap = map[int8]string{
	0: "待发货",
	1: "已发货",
	2: "运输中",
	3: "已送达",
	4: "异常",
}

// 偏远地区（运费加价）
var remoteAreas = map[string]bool{
	"新疆": true, "西藏": true, "内蒙古": true, "青海": true, "甘肃": true, "宁夏": true,
}

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

// companyName 根据编码返回公司名
func companyName(code string) string {
	if name, ok := companyNameMap[code]; ok {
		return name
	}
	return code
}

// generateLogisticsNo 生成物流单号：公司码 + 时间戳 + 6位随机
func generateLogisticsNo(companyCode string) string {
	if companyCode == "" {
		companyCode = "EXP"
	}
	return fmt.Sprintf("%s%s%06d", companyCode, time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// simulateLocation 根据状态返回模拟位置（demo 未接第三方物流 API）
func simulateLocation(status int8) string {
	switch status {
	case 1:
		return "上海转运中心"
	case 2:
		return "杭州转运中心"
	case 3:
		return "目的地"
	default:
		return ""
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
	logisticsNo := generateLogisticsNo(req.CompanyCode)
	if req.CompanyName == "" {
		req.CompanyName = companyName(req.CompanyCode)
	}
	now := time.Now()

	logistics := &model.Logistics{
		OrderID:          req.OrderID,
		OrderNo:          req.OrderNo,
		LogisticsCompany: req.CompanyName,
		LogisticsNo:      logisticsNo,
		ReceiverName:     req.ReceiverName,
		ReceiverPhone:    req.ReceiverPhone,
		ReceiverAddress:  req.ReceiverAddress,
		Status:           0, // 待发货
		TrackingInfo: model.TrackingList{
			{Time: now.Format(time.RFC3339), Status: "待发货", Remark: "物流单已创建"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := l.logisticsRepo.Create(ctx, logistics); err != nil {
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
	Location    string
}

// UpdateLogisticsStatus 更新物流状态（同时追加轨迹节点）
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

	location := req.Location
	if location == "" {
		location = simulateLocation(req.Status)
	}
	if location != "" {
		logistics.CurrentLocation = &location
	}

	remark := req.Remark
	if remark == "" {
		remark = statusTextMap[req.Status]
	}

	// 追加轨迹节点
	logistics.TrackingInfo = append(logistics.TrackingInfo, model.TrackingNode{
		Time:     now.Format(time.RFC3339),
		Status:   statusTextMap[req.Status],
		Location: location,
		Remark:   remark,
	})

	if err := l.logisticsRepo.Update(ctx, logistics); err != nil {
		return apperrors.NewInternalError("更新物流状态失败")
	}

	return nil
}

// QueryTrackingRequest 查询物流轨迹请求
type QueryTrackingRequest struct {
	LogisticsNo string
}

// QueryTrackingResponse 查询物流轨迹响应
type QueryTrackingResponse struct {
	Nodes []*model.TrackingNode
}

// QueryTracking 查询物流轨迹（返回真实累积的轨迹节点）
func (l *LogisticsLogic) QueryTracking(ctx context.Context, req *QueryTrackingRequest) (*QueryTrackingResponse, error) {
	logistics, err := l.logisticsRepo.GetByLogisticsNo(ctx, req.LogisticsNo)
	if err != nil {
		return nil, apperrors.NewNotFoundError("物流信息不存在")
	}

	nodes := make([]*model.TrackingNode, 0, len(logistics.TrackingInfo))
	for i := range logistics.TrackingInfo {
		n := logistics.TrackingInfo[i]
		nodes = append(nodes, &model.TrackingNode{
			Time:     n.Time,
			Status:   n.Status,
			Location: n.Location,
			Remark:   n.Remark,
		})
	}

	// 兜底：无轨迹时至少返回创建节点
	if len(nodes) == 0 {
		nodes = append(nodes, &model.TrackingNode{
			Time:   logistics.CreatedAt.Format(time.RFC3339),
			Status: "待发货",
			Remark: "物流单已创建",
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

// CalculateFreight 计算运费：基础费 + 超重 + 体积 + 偏远加价
func (l *LogisticsLogic) CalculateFreight(ctx context.Context, req *CalculateFreightRequest) (*CalculateFreightResponse, error) {
	base := 8.0
	// 首重 1kg 内免超重费，超出部分每 kg 2 元
	if req.Weight > 1.0 {
		base += (req.Weight - 1.0) * 2.0
	}
	// 体积超过 0.03m³ 部分，每 0.01m³ 加 0.5 元
	if req.Volume > 0.03 {
		base += (req.Volume - 0.03) / 0.01 * 0.5
	}
	// 偏远地区加价
	if remoteAreas[req.Province] {
		base += 5.0
	}

	freight := math.Round(base*100) / 100
	if freight < 6 {
		freight = 6
	}

	return &CalculateFreightResponse{
		Freight: freight,
	}, nil
}

// ListLogisticsRequest 获取物流列表请求
type ListLogisticsRequest struct {
	Page     int
	PageSize int
	OrderNo  string
}

// ListLogisticsResponse 获取物流列表响应
type ListLogisticsResponse struct {
	Logistics []*model.Logistics
	Total     int64
}

// ListLogistics 获取物流列表（分页，可选按订单号过滤）
func (l *LogisticsLogic) ListLogistics(ctx context.Context, req *ListLogisticsRequest) (*ListLogisticsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	list, total, err := l.logisticsRepo.List(ctx, req.Page, req.PageSize, req.OrderNo)
	if err != nil {
		return nil, apperrors.NewInternalError("查询物流列表失败")
	}

	return &ListLogisticsResponse{
		Logistics: list,
		Total:     total,
	}, nil
}
