package job

import (
	"context"

	v1 "github.com/zhian9/GoForge/server/api/job/v1"
	"github.com/zhian9/GoForge/server/internal/service/job/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// JobService 实现 gRPC 服务接口
type JobService struct {
	v1.UnimplementedJobServiceServer
	svcCtx *ServiceContext
	logic  *service.JobLogic
}

// NewJobService 创建定时任务服务
func NewJobService(svcCtx *ServiceContext) *JobService {
	logic := service.NewJobLogic(
		svcCtx.OrderRepo,
		svcCtx.CouponRepo,
	)

	return &JobService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// CancelExpiredOrders 订单超时取消
func (s *JobService) CancelExpiredOrders(ctx context.Context, req *v1.CancelExpiredOrdersRequest) (*v1.CancelExpiredOrdersResponse, error) {
	cancelReq := &service.CancelExpiredOrdersRequest{
		TimeoutMinutes: int(req.TimeoutMinutes),
	}

	resp, err := s.logic.CancelExpiredOrders(ctx, cancelReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.CancelExpiredOrdersResponse{
		Code:           0,
		Message:        "处理成功",
		CancelledCount: int32(resp.CancelledCount),
	}, nil
}

// ProcessExpiredCoupons 优惠券过期处理
func (s *JobService) ProcessExpiredCoupons(ctx context.Context, req *v1.ProcessExpiredCouponsRequest) (*v1.ProcessExpiredCouponsResponse, error) {
	processReq := &service.ProcessExpiredCouponsRequest{}

	resp, err := s.logic.ProcessExpiredCoupons(ctx, processReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ProcessExpiredCouponsResponse{
		Code:         0,
		Message:      "处理成功",
		ExpiredCount: int32(resp.ExpiredCount),
	}, nil
}

// GenerateStatistics 数据统计
func (s *JobService) GenerateStatistics(ctx context.Context, req *v1.GenerateStatisticsRequest) (*v1.GenerateStatisticsResponse, error) {
	genReq := &service.GenerateStatisticsRequest{
		Date: req.Date,
	}

	err := s.logic.GenerateStatistics(ctx, genReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GenerateStatisticsResponse{
		Code:    0,
		Message: "统计成功",
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}
