package file

import (
	"context"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/file/v1"
	"github.com/zhian9/GoForge/server/internal/service/file/model"
	"github.com/zhian9/GoForge/server/internal/service/file/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FileService 实现 gRPC 服务接口
type FileService struct {
	v1.UnimplementedFileServiceServer
	svcCtx *ServiceContext
	logic  *service.FileLogic
}

// NewFileService 创建文件服务
func NewFileService(svcCtx *ServiceContext) *FileService {
	logic := service.NewFileLogic(svcCtx.FileRepo)

	return &FileService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// UploadFile 上传文件
func (s *FileService) UploadFile(ctx context.Context, req *v1.UploadFileRequest) (*v1.UploadFileResponse, error) {
	uploadReq := &service.UploadFileRequest{
		FileData: req.FileData,
		FileName: req.FileName,
		FileType: req.FileType,
		Category: req.Category,
	}

	resp, err := s.logic.UploadFile(ctx, uploadReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.UploadFileResponse{
		Code:    0,
		Message: "上传成功",
		Data:    convertFileInfoToProto(resp.FileInfo),
	}, nil
}

// BatchUploadFile 批量上传文件
func (s *FileService) BatchUploadFile(ctx context.Context, req *v1.BatchUploadFileRequest) (*v1.BatchUploadFileResponse, error) {
	batchReq := &service.BatchUploadFileRequest{
		FileDataList: req.FileData,
		FileNames:    req.FileNames,
		Category:     req.Category,
	}

	resp, err := s.logic.BatchUploadFile(ctx, batchReq)
	if err != nil {
		return nil, convertError(err)
	}

	fileInfos := make([]*v1.FileInfo, 0, len(resp.FileInfos))
	for _, fi := range resp.FileInfos {
		fileInfos = append(fileInfos, convertFileInfoToProto(fi))
	}

	return &v1.BatchUploadFileResponse{
		Code:    0,
		Message: "上传成功",
		Data:    fileInfos,
	}, nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(ctx context.Context, req *v1.DeleteFileRequest) (*v1.DeleteFileResponse, error) {
	deleteReq := &service.DeleteFileRequest{
		FileID: req.FileId,
	}

	err := s.logic.DeleteFile(ctx, deleteReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.DeleteFileResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// GetFileURL 获取文件URL
func (s *FileService) GetFileURL(ctx context.Context, req *v1.GetFileURLRequest) (*v1.GetFileURLResponse, error) {
	getReq := &service.GetFileURLRequest{
		FileID: req.FileId,
	}

	resp, err := s.logic.GetFileURL(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetFileURLResponse{
		Code:    0,
		Message: "成功",
		FileUrl: resp.FileURL,
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}

// convertFileInfoToProto 转换文件信息模型为 Protobuf 消息
func convertFileInfoToProto(fi *model.FileInfo) *v1.FileInfo {
	if fi == nil {
		return nil
	}
	return &v1.FileInfo{
		FileId:    fi.FileID,
		FileName:  fi.FileName,
		FileUrl:   fi.FileURL,
		FileSize:  fi.FileSize,
		FileType:  fi.FileType,
		CreatedAt: formatTime(&fi.CreatedAt),
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
