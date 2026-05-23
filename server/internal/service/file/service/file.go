package service

import (
	"context"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/file/model"
	"github.com/zhian9/GoForge/server/internal/service/file/repository"
)

// FileLogic 文件业务逻辑
type FileLogic struct {
	fileRepo repository.FileRepository
}

// NewFileLogic 创建文件业务逻辑
func NewFileLogic(fileRepo repository.FileRepository) *FileLogic {
	return &FileLogic{
		fileRepo: fileRepo,
	}
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	FileData []byte
	FileName string
	FileType string
	Category string
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	FileInfo *model.FileInfo
}

// UploadFile 上传文件
func (l *FileLogic) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	if len(req.FileData) == 0 {
		return nil, apperrors.NewInvalidParamError("文件数据不能为空")
	}

	fileInfo, err := l.fileRepo.UploadFile(ctx, req.FileData, req.FileName, req.FileType, req.Category)
	if err != nil {
		return nil, apperrors.NewInternalError("上传文件失败")
	}

	return &UploadFileResponse{
		FileInfo: fileInfo,
	}, nil
}

// BatchUploadFileRequest 批量上传文件请求
type BatchUploadFileRequest struct {
	FileDataList [][]byte
	FileNames    []string
	Category     string
}

// BatchUploadFileResponse 批量上传文件响应
type BatchUploadFileResponse struct {
	FileInfos []*model.FileInfo
}

// BatchUploadFile 批量上传文件
func (l *FileLogic) BatchUploadFile(ctx context.Context, req *BatchUploadFileRequest) (*BatchUploadFileResponse, error) {
	fileInfos := make([]*model.FileInfo, 0, len(req.FileDataList))

	for i, fileData := range req.FileDataList {
		if i < len(req.FileNames) {
			fileInfo, err := l.fileRepo.UploadFile(ctx, fileData, req.FileNames[i], "", req.Category)
			if err == nil {
				fileInfos = append(fileInfos, fileInfo)
			}
		}
	}

	return &BatchUploadFileResponse{
		FileInfos: fileInfos,
	}, nil
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	FileID string
}

// DeleteFile 删除文件
func (l *FileLogic) DeleteFile(ctx context.Context, req *DeleteFileRequest) error {
	err := l.fileRepo.DeleteFile(ctx, req.FileID)
	if err != nil {
		return apperrors.NewInternalError("删除文件失败")
	}
	return nil
}

// GetFileURLRequest 获取文件URL请求
type GetFileURLRequest struct {
	FileID string
}

// GetFileURLResponse 获取文件URL响应
type GetFileURLResponse struct {
	FileURL string
}

// GetFileURL 获取文件URL
func (l *FileLogic) GetFileURL(ctx context.Context, req *GetFileURLRequest) (*GetFileURLResponse, error) {
	fileURL, err := l.fileRepo.GetFileURL(ctx, req.FileID)
	if err != nil {
		return nil, apperrors.NewInternalError("获取文件URL失败")
	}

	return &GetFileURLResponse{
		FileURL: fileURL,
	}, nil
}
