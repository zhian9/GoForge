package model

import (
	"time"
)

// FileInfo 文件信息模型
type FileInfo struct {
	FileID    string
	FileName  string
	FileURL   string
	FileSize  int64
	FileType  string
	CreatedAt time.Time
}
