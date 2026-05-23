package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "github.com/zhian9/GoForge/server/api/file/v1"
)

// FileUploadHandler 文件上传处理器
type FileUploadHandler struct {
	fileServiceClient v1.FileServiceClient
	conn              *grpc.ClientConn
}

// apiResp 统一 HTTP JSON 响应结构（避免 proto 的 `omitempty` 导致 code=0 被省略）
type apiResp[T any] struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

// NewFileUploadHandler 创建文件上传处理器
func NewFileUploadHandler(fileServiceAddr string) (*FileUploadHandler, error) {
	conn, err := grpc.NewClient(
		fileServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("连接文件服务失败: %w", err)
	}

	client := v1.NewFileServiceClient(conn)

	return &FileUploadHandler{
		fileServiceClient: client,
		conn:              conn,
	}, nil
}

// Close 关闭连接
func (h *FileUploadHandler) Close() error {
	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}

// HandleUpload 处理文件上传
func (h *FileUploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Content-Length")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 只处理 POST 请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制请求体大小（100MB，增加限制避免 413 错误）
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB

	// 解析 multipart/form-data（增加限制）
	err := r.ParseMultipartForm(100 << 20) // 100MB
	if err != nil {
		http.Error(w, fmt.Sprintf("解析表单失败: %v", err), http.StatusBadRequest)
		return
	}

	// 获取文件
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("获取文件失败: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取文件失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取 category（可选）
	category := r.FormValue("category")
	if category == "" {
		category = "image" // 默认分类
	}

	// 获取文件类型
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// 根据文件扩展名推断
		fileName := header.Filename
		if strings.HasSuffix(strings.ToLower(fileName), ".jpg") || strings.HasSuffix(strings.ToLower(fileName), ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(strings.ToLower(fileName), ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(strings.ToLower(fileName), ".gif") {
			contentType = "image/gif"
		} else {
			contentType = "application/octet-stream"
		}
	}

	// 调用 gRPC 服务上传文件
	ctx := context.Background()
	req := &v1.UploadFileRequest{
		FileData: fileData,
		FileName: header.Filename,
		FileType: contentType,
		Category: category,
	}

	resp, err := h.fileServiceClient.UploadFile(ctx, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("上传文件失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apiResp[*v1.FileInfo]{
		Code:    resp.GetCode(),
		Message: resp.GetMessage(),
		Data:    resp.GetData(),
	})
}

// HandleBatchUpload 处理批量文件上传
func (h *FileUploadHandler) HandleBatchUpload(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Content-Length")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 只处理 POST 请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制请求体大小（200MB，批量上传需要更大空间）
	r.Body = http.MaxBytesReader(w, r.Body, 200<<20) // 200MB

	// 解析 multipart/form-data（增加限制）
	err := r.ParseMultipartForm(200 << 20) // 200MB
	if err != nil {
		http.Error(w, fmt.Sprintf("解析表单失败: %v", err), http.StatusBadRequest)
		return
	}

	// 获取文件列表
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "没有上传文件", http.StatusBadRequest)
		return
	}

	// 获取 category（可选）
	category := r.FormValue("category")
	if category == "" {
		category = "image" // 默认分类
	}

	var fileDataList [][]byte
	var fileNames []string

	// 读取所有文件
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, fmt.Sprintf("打开文件失败: %v", err), http.StatusInternalServerError)
			return
		}

		fileData, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			http.Error(w, fmt.Sprintf("读取文件失败: %v", err), http.StatusInternalServerError)
			return
		}

		fileDataList = append(fileDataList, fileData)
		fileNames = append(fileNames, fileHeader.Filename)
	}

	// 调用 gRPC 服务批量上传文件
	ctx := context.Background()
	req := &v1.BatchUploadFileRequest{
		FileData:  fileDataList,
		FileNames: fileNames,
		Category:  category,
	}

	resp, err := h.fileServiceClient.BatchUploadFile(ctx, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("批量上传文件失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apiResp[[]*v1.FileInfo]{
		Code:    resp.GetCode(),
		Message: resp.GetMessage(),
		Data:    resp.GetData(),
	})
}
