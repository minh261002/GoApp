package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go_app/configs"
	"go_app/pkg/logger"
	"go_app/pkg/utils"
)

type UploadService struct {
	config *configs.Config
}

func NewUploadService() *UploadService {
	return &UploadService{
		config: configs.Load(),
	}
}

// UploadFileRequest represents the request to upload a file
type UploadFileRequest struct {
	File        *multipart.FileHeader `form:"file" binding:"required"`
	Folder      string                `form:"folder"`
	AllowedExts []string              `form:"-"`
	MaxSize     int64                 `form:"-"`
}

// UploadFileResponse represents the response for file upload
type UploadFileResponse struct {
	FileName     string    `json:"file_name"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"file_path"`
	FileURL      string    `json:"file_url"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// UploadFile uploads a file to the server
func (s *UploadService) UploadFile(req *UploadFileRequest) (*UploadFileResponse, error) {
	// Set default values
	if req.MaxSize == 0 {
		req.MaxSize = 10 * 1024 * 1024 // 10MB default
	}
	if req.Folder == "" {
		req.Folder = "uploads"
	}
	if len(req.AllowedExts) == 0 {
		req.AllowedExts = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx"}
	}

	// Validate file size
	if req.File.Size > req.MaxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", req.MaxSize)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(req.File.Filename))
	if ext == "" {
		ext = ".bin"
	}

	// Validate file extension
	allowed := false
	for _, allowedExt := range req.AllowedExts {
		if ext == strings.ToLower(allowedExt) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("file extension %s is not allowed. Allowed extensions: %s", ext, strings.Join(req.AllowedExts, ", "))
	}

	// Generate unique filename
	fileName := s.generateUniqueFileName(req.File.Filename)

	// Create upload directory
	uploadDir := filepath.Join("uploads", req.Folder)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create file path
	filePath := filepath.Join(uploadDir, fileName)

	// Open uploaded file
	src, err := req.File.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Verify file was written completely
	if written != req.File.Size {
		return nil, fmt.Errorf("file size mismatch: expected %d, got %d", req.File.Size, written)
	}

	// Generate file URL
	fileURL := s.generateFileURL(req.Folder, fileName)

	response := &UploadFileResponse{
		FileName:     fileName,
		OriginalName: req.File.Filename,
		FilePath:     filePath,
		FileURL:      fileURL,
		FileSize:     req.File.Size,
		MimeType:     req.File.Header.Get("Content-Type"),
		UploadedAt:   time.Now(),
	}

	logger.Infof("File uploaded successfully: %s", filePath)
	return response, nil
}

// UploadMultipleFiles uploads multiple files
func (s *UploadService) UploadMultipleFiles(files []*multipart.FileHeader, folder string, allowedExts []string, maxSize int64) ([]*UploadFileResponse, error) {
	var responses []*UploadFileResponse
	var errors []string

	for _, file := range files {
		req := &UploadFileRequest{
			File:        file,
			Folder:      folder,
			AllowedExts: allowedExts,
			MaxSize:     maxSize,
		}

		response, err := s.UploadFile(req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %s: %v", file.Filename, err))
			continue
		}

		responses = append(responses, response)
	}

	if len(errors) > 0 && len(responses) == 0 {
		return nil, fmt.Errorf("all uploads failed: %s", strings.Join(errors, "; "))
	}

	return responses, nil
}

// DeleteFile deletes a file from the server
func (s *UploadService) DeleteFile(filePath string) error {
	// Security check: ensure file is within uploads directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	uploadsDir, err := filepath.Abs("uploads")
	if err != nil {
		return fmt.Errorf("failed to get uploads directory: %w", err)
	}

	if !strings.HasPrefix(absPath, uploadsDir) {
		return fmt.Errorf("file path is outside uploads directory")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Infof("File deleted successfully: %s", filePath)
	return nil
}

// GetFileInfo gets information about a file
func (s *UploadService) GetFileInfo(filePath string) (*UploadFileResponse, error) {
	// Security check
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	uploadsDir, err := filepath.Abs("uploads")
	if err != nil {
		return nil, fmt.Errorf("failed to get uploads directory: %w", err)
	}

	if !strings.HasPrefix(absPath, uploadsDir) {
		return nil, fmt.Errorf("file path is outside uploads directory")
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Generate file URL
	relPath, err := filepath.Rel("uploads", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	fileURL := s.generateFileURL(filepath.Dir(relPath), filepath.Base(relPath))

	response := &UploadFileResponse{
		FileName:     fileInfo.Name(),
		OriginalName: fileInfo.Name(),
		FilePath:     filePath,
		FileURL:      fileURL,
		FileSize:     fileInfo.Size(),
		MimeType:     s.getMimeType(filePath),
		UploadedAt:   fileInfo.ModTime(),
	}

	return response, nil
}

// generateUniqueFileName generates a unique filename
func (s *UploadService) generateUniqueFileName(originalName string) string {
	// Get file extension
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".bin"
	}

	// Generate hash from original name + timestamp
	timestamp := time.Now().UnixNano()
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%d", originalName, timestamp)))
	hashStr := fmt.Sprintf("%x", hash)[:8]

	// Create unique filename
	baseName := strings.TrimSuffix(originalName, ext)
	baseName = utils.GenerateSlug(baseName)
	if baseName == "" {
		baseName = "file"
	}

	return fmt.Sprintf("%s_%s%s", baseName, hashStr, ext)
}

// generateFileURL generates a file URL
func (s *UploadService) generateFileURL(folder, fileName string) string {
	baseURL := s.config.Server.Host
	if s.config.Server.Port != "80" && s.config.Server.Port != "443" {
		baseURL = fmt.Sprintf("%s:%s", baseURL, s.config.Server.Port)
	}

	// Use HTTP for now, in production should use HTTPS
	protocol := "http"
	if s.config.Server.Port == "443" {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s/uploads/%s/%s", protocol, baseURL, folder, fileName)
}

// getMimeType gets the MIME type of a file
func (s *UploadService) getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".json": "application/json",
		".xml":  "application/xml",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// ValidateImageFile validates if a file is a valid image
func (s *UploadService) ValidateImageFile(file *multipart.FileHeader) error {
	// Check file size (max 5MB for images)
	if file.Size > 5*1024*1024 {
		return fmt.Errorf("image file size exceeds maximum allowed size of 5MB")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedImageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	allowed := false
	for _, allowedExt := range allowedImageExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("image file extension %s is not allowed. Allowed extensions: %s", ext, strings.Join(allowedImageExts, ", "))
	}

	// Check MIME type
	mimeType := file.Header.Get("Content-Type")
	allowedMimeTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	allowed = false
	for _, allowedMimeType := range allowedMimeTypes {
		if mimeType == allowedMimeType {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("image MIME type %s is not allowed. Allowed types: %s", mimeType, strings.Join(allowedMimeTypes, ", "))
	}

	return nil
}
