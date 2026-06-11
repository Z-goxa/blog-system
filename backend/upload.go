package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadHandler 处理上传文件的 HTTP 请求
type UploadHandler struct {
	UploadDir string
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/uploads/") {
		filePath := filepath.Join(h.UploadDir, strings.TrimPrefix(r.URL.Path, "/uploads/"))
		http.ServeFile(w, r, filePath)
		return
	}
	http.NotFound(w, r)
}

// UploadService 上传服务
type UploadService struct {
	uploadDir string
}

// NewUploadService 创建上传服务
func NewUploadService() *UploadService {
	dir := "uploads"
	os.MkdirAll(dir, 0755)
	return &UploadService{
		uploadDir: dir,
	}
}

// UploadResult 上传结果
type UploadResult struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// UploadImage 上传图片（从 Base64 或文件路径）
func (s *UploadService) UploadImage(filename string, data []byte) (*UploadResult, error) {
	// 生成唯一文件名
	ext := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(s.uploadDir, newFilename)

	// 保存文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	log.Printf("✅ 图片上传成功: %s", filePath)

	return &UploadResult{
		URL:  "/uploads/" + newFilename,
		Name: filename,
		Size: int64(len(data)),
	}, nil
}

// GetUploadDir 获取上传目录（供前端使用）
func (s *UploadService) GetUploadDir() string {
	absPath, _ := filepath.Abs(s.uploadDir)
	return absPath
}
