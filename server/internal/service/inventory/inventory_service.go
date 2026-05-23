package inventory

import (
	"context"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/inventory/v1"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/inventory/model"
	"github.com/zhian9/GoForge/server/internal/service/inventory/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InventoryService 实现 gRPC 服务接口
type InventoryService struct {
	v1.UnimplementedInventoryServiceServer
	svcCtx *ServiceContext
	logic  *service.InventoryLogic
}

// NewInventoryService 创建库存服务
func NewInventoryService(svcCtx *ServiceContext) *InventoryService {
	logic := service.NewInventoryLogic(
		svcCtx.InventoryRepo,
		svcCtx.InventoryLogRepo,
		svcCtx.Cache,
		svcCtx.MQProducer,
	)

	return &InventoryService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// GetInventory 获取库存
func (s *InventoryService) GetInventory(ctx context.Context, req *v1.GetInventoryRequest) (*v1.GetInventoryResponse, error) {
	getReq := &service.GetInventoryRequest{
		SkuID: uint64(req.SkuId),
	}

	resp, err := s.logic.GetInventory(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetInventoryResponse{
		Code:    0,
		Message: "成功",
		Data:    convertInventoryToProto(resp.Inventory),
	}, nil
}

// BatchGetInventory 批量获取库存
func (s *InventoryService) BatchGetInventory(ctx context.Context, req *v1.BatchGetInventoryRequest) (*v1.BatchGetInventoryResponse, error) {
	skuIDs := make([]uint64, 0, len(req.SkuIds))
	for _, id := range req.SkuIds {
		skuIDs = append(skuIDs, uint64(id))
	}

	getReq := &service.BatchGetInventoryRequest{
		SkuIDs: skuIDs,
	}

	resp, err := s.logic.BatchGetInventory(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	inventories := make([]*v1.Inventory, 0, len(resp.Inventories))
	for _, inv := range resp.Inventories {
		if inv != nil {
			inventories = append(inventories, convertInventoryToProto(inv))
		}
	}

	return &v1.BatchGetInventoryResponse{
		Code:    0,
		Message: "成功",
		Data:    inventories,
	}, nil
}

// LockStock 锁定库存
func (s *InventoryService) LockStock(ctx context.Context, req *v1.LockStockRequest) (*v1.LockStockResponse, error) {
	lockReq := &service.LockStockRequest{
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
		OrderID:  uint64(req.OrderId),
		Remark:   req.Remark,
	}

	err := s.logic.LockStock(ctx, lockReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.LockStockResponse{
		Code:    0,
		Message: "锁定成功",
	}, nil
}

// DeductStock 扣减库存
func (s *InventoryService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.DeductStockResponse, error) {
	deductReq := &service.DeductStockRequest{
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
		OrderID:  uint64(req.OrderId),
		Remark:   req.Remark,
	}

	err := s.logic.DeductStock(ctx, deductReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.DeductStockResponse{
		Code:    0,
		Message: "扣减成功",
	}, nil
}

// UnlockStock 解锁库存
func (s *InventoryService) UnlockStock(ctx context.Context, req *v1.UnlockStockRequest) (*v1.UnlockStockResponse, error) {
	unlockReq := &service.UnlockStockRequest{
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
		OrderID:  uint64(req.OrderId),
		Remark:   req.Remark,
	}

	err := s.logic.UnlockStock(ctx, unlockReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.UnlockStockResponse{
		Code:    0,
		Message: "解锁成功",
	}, nil
}

// RollbackStock 回退库存
func (s *InventoryService) RollbackStock(ctx context.Context, req *v1.RollbackStockRequest) (*v1.RollbackStockResponse, error) {
	rollbackReq := &service.RollbackStockRequest{
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
		OrderID:  uint64(req.OrderId),
		Remark:   req.Remark,
	}

	err := s.logic.RollbackStock(ctx, rollbackReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.RollbackStockResponse{
		Code:    0,
		Message: "回退成功",
	}, nil
}

// StockIn 入库
func (s *InventoryService) StockIn(ctx context.Context, req *v1.StockInRequest) (*v1.StockInResponse, error) {
	stockInReq := &service.StockInRequest{
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
		Remark:   req.Remark,
	}

	err := s.logic.StockIn(ctx, stockInReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.StockInResponse{
		Code:    0,
		Message: "入库成功",
	}, nil
}

// GetInventoryLog 获取库存流水
func (s *InventoryService) GetInventoryLog(ctx context.Context, req *v1.GetInventoryLogRequest) (*v1.GetInventoryLogResponse, error) {
	getReq := &service.GetInventoryLogRequest{
		SkuID:    uint64(req.SkuId),
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	resp, err := s.logic.GetInventoryLog(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	logs := make([]*v1.InventoryLog, 0, len(resp.Logs))
	for _, log := range resp.Logs {
		if log != nil {
			logs = append(logs, convertInventoryLogToProto(log))
		}
	}

	return &v1.GetInventoryLogResponse{
		Code:    0,
		Message: "成功",
		Data:    logs,
		Total:   int32(resp.Total),
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
		case apperrors.CodeNotFound:
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

// convertInventoryToProto 转换库存模型为 Protobuf 消息
func convertInventoryToProto(inv *model.Inventory) *v1.Inventory {
	if inv == nil {
		return nil
	}

	return &v1.Inventory{
		Id:                int64(inv.ID),
		SkuId:             int64(inv.SkuID),
		TotalStock:        int32(inv.TotalStock),
		AvailableStock:    int32(inv.AvailableStock),
		LockedStock:       int32(inv.LockedStock),
		SoldStock:         int32(inv.SoldStock),
		LowStockThreshold: int32(inv.LowStockThreshold),
		CreatedAt:         formatTime(&inv.CreatedAt),
		UpdatedAt:         formatTime(&inv.UpdatedAt),
	}
}

// convertInventoryLogToProto 转换库存流水模型为 Protobuf 消息
func convertInventoryLogToProto(log *model.InventoryLog) *v1.InventoryLog {
	if log == nil {
		return nil
	}

	var orderID int64
	if log.OrderID != nil {
		orderID = int64(*log.OrderID)
	}

	return &v1.InventoryLog{
		Id:          int64(log.ID),
		SkuId:       int64(log.SkuID),
		OrderId:     orderID,
		Type:        int32(log.Type),
		Quantity:    int32(log.Quantity),
		BeforeStock: int32(log.BeforeStock),
		AfterStock:  int32(log.AfterStock),
		Remark:      log.Remark,
		CreatedAt:   formatTime(&log.CreatedAt),
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
