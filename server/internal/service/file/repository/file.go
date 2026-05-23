package repository

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zhian9/GoForge/server/internal/service/file/model"
)

// FileRepository 文件仓库接口
type FileRepository interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, fileData []byte, fileName, fileType, category string) (*model.FileInfo, error)
	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, fileID string) error
	// GetFileURL 获取文件URL
	GetFileURL(ctx context.Context, fileID string) (string, error)
}

type fileRepository struct {
	storageType string
	localPath   string
}

// NewFileRepository 创建文件仓库
func NewFileRepository(storageConfig interface{}) FileRepository {
	// 简化处理，实际应该解析配置
	return &fileRepository{
		storageType: "local",
		localPath:   "./uploads",
	}
}

// generateFileID 生成文件ID
func (r *fileRepository) generateFileID(fileData []byte, fileName string) string {
	h := md5.New()
	h.Write(fileData)
	h.Write([]byte(fileName))
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}

// UploadFile 上传文件
func (r *fileRepository) UploadFile(ctx context.Context, fileData []byte, fileName, fileType, category string) (*model.FileInfo, error) {
	fileID := r.generateFileID(fileData, fileName)

	// 创建目录
	dir := filepath.Join(r.localPath, category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}

	// 保存文件
	filePath := filepath.Join(dir, fileID+filepath.Ext(fileName))
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(fileData); err != nil {
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}

	fileInfo := &model.FileInfo{
		FileID:    fileID,
		FileName:  fileName,
		FileURL:   "/uploads/" + category + "/" + fileID + filepath.Ext(fileName),
		FileSize:  int64(len(fileData)),
		FileType:  fileType,
		CreatedAt: time.Now(),
	}

	return fileInfo, nil
}

// DeleteFile 删除文件
func (r *fileRepository) DeleteFile(ctx context.Context, fileID string) error {
	// 简化处理，实际应该根据fileID查找文件路径
	// 这里只是示例
	return nil
}

// GetFileURL 获取文件URL
func (r *fileRepository) GetFileURL(ctx context.Context, fileID string) (string, error) {
	// 简化处理，实际应该根据fileID查找文件URL
	return "/uploads/" + fileID, nil
}
