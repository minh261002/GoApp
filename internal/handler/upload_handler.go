package handler

import (
	"net/http"

	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{
		uploadService: service.NewUploadService(),
	}
}

// UploadFile uploads a single file
// @Summary Upload a single file
// @Description Upload a single file to the server
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param folder formData string false "Folder to upload to" default(uploads)
// @Success 200 {object} response.SuccessResponse{data=service.UploadFileResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 413 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload [post]
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// Parse form data
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
		return
	}

	folder := c.PostForm("folder")
	if folder == "" {
		folder = "general"
	}

	// Create upload request
	req := &service.UploadFileRequest{
		File:        file,
		Folder:      folder,
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv"},
		MaxSize:     10 * 1024 * 1024, // 10MB
	}

	// Upload file
	result, err := h.uploadService.UploadFile(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to upload file", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "File uploaded successfully", result)
}

// UploadImage uploads an image file
// @Summary Upload an image file
// @Description Upload an image file with validation
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file to upload"
// @Param folder formData string false "Folder to upload to" default(images)
// @Success 200 {object} response.SuccessResponse{data=service.UploadFileResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 413 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Parse form data
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
		return
	}

	// Validate image file
	if err := h.uploadService.ValidateImageFile(file); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid image file", err.Error())
		return
	}

	folder := c.PostForm("folder")
	if folder == "" {
		folder = "images"
	}

	// Create upload request for image
	req := &service.UploadFileRequest{
		File:        file,
		Folder:      folder,
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		MaxSize:     5 * 1024 * 1024, // 5MB for images
	}

	// Upload file
	result, err := h.uploadService.UploadFile(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to upload image", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Image uploaded successfully", result)
}

// UploadBrandLogo uploads a brand logo
// @Summary Upload brand logo
// @Description Upload a brand logo image
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Brand logo image"
// @Success 200 {object} response.SuccessResponse{data=service.UploadFileResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 413 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload/brand-logo [post]
func (h *UploadHandler) UploadBrandLogo(c *gin.Context) {
	// Parse form data
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
		return
	}

	// Validate image file
	if err := h.uploadService.ValidateImageFile(file); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid image file", err.Error())
		return
	}

	// Create upload request for brand logo
	req := &service.UploadFileRequest{
		File:        file,
		Folder:      "brands/logos",
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		MaxSize:     2 * 1024 * 1024, // 2MB for brand logos
	}

	// Upload file
	result, err := h.uploadService.UploadFile(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to upload brand logo", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brand logo uploaded successfully", result)
}

// UploadMultipleFiles uploads multiple files
// @Summary Upload multiple files
// @Description Upload multiple files at once
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Files to upload"
// @Param folder formData string false "Folder to upload to" default(uploads)
// @Success 200 {object} response.SuccessResponse{data=[]service.UploadFileResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 413 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload/multiple [post]
func (h *UploadHandler) UploadMultipleFiles(c *gin.Context) {
	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to parse form data", err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.Error(c, http.StatusBadRequest, "No files provided", "At least one file is required")
		return
	}

	if len(files) > 10 {
		response.Error(c, http.StatusBadRequest, "Too many files", "Maximum 10 files allowed")
		return
	}

	folder := c.PostForm("folder")
	if folder == "" {
		folder = "general"
	}

	// Upload files
	results, err := h.uploadService.UploadMultipleFiles(
		files,
		folder,
		[]string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv"},
		10*1024*1024, // 10MB per file
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to upload files", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Files uploaded successfully", results)
}

// DeleteFile deletes a file
// @Summary Delete a file
// @Description Delete a file from the server
// @Tags upload
// @Accept json
// @Produce json
// @Param file_path query string true "File path to delete"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload [delete]
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	filePath := c.Query("file_path")
	if filePath == "" {
		response.Error(c, http.StatusBadRequest, "File path is required", "file_path parameter is required")
		return
	}

	// Delete file
	err := h.uploadService.DeleteFile(filePath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete file", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "File deleted successfully", nil)
}

// GetFileInfo gets file information
// @Summary Get file information
// @Description Get information about a file
// @Tags upload
// @Accept json
// @Produce json
// @Param file_path query string true "File path"
// @Success 200 {object} response.SuccessResponse{data=service.UploadFileResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload/info [get]
func (h *UploadHandler) GetFileInfo(c *gin.Context) {
	filePath := c.Query("file_path")
	if filePath == "" {
		response.Error(c, http.StatusBadRequest, "File path is required", "file_path parameter is required")
		return
	}

	// Get file info
	info, err := h.uploadService.GetFileInfo(filePath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get file info", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "File info retrieved successfully", info)
}

// UploadDocument uploads a document file
// @Summary Upload a document file
// @Description Upload a document file (PDF, DOC, DOCX, TXT, CSV)
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Document file to upload"
// @Param folder formData string false "Folder to upload to" default(documents)
// @Success 200 {object} response.SuccessResponse{data=service.UploadFileResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 413 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload/document [post]
func (h *UploadHandler) UploadDocument(c *gin.Context) {
	// Parse form data
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
		return
	}

	folder := c.PostForm("folder")
	if folder == "" {
		folder = "documents"
	}

	// Create upload request for document
	req := &service.UploadFileRequest{
		File:        file,
		Folder:      folder,
		AllowedExts: []string{".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx", ".xls"},
		MaxSize:     20 * 1024 * 1024, // 20MB for documents
	}

	// Upload file
	result, err := h.uploadService.UploadFile(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to upload document", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Document uploaded successfully", result)
}

// GetUploadStats gets upload statistics
// @Summary Get upload statistics
// @Description Get statistics about uploaded files
// @Tags upload
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=map[string]interface{}}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/upload/stats [get]
func (h *UploadHandler) GetUploadStats(c *gin.Context) {
	// This would typically query the database for upload statistics
	// For now, return a simple response
	stats := map[string]interface{}{
		"total_uploads": 0,
		"total_size":    "0 MB",
		"file_types": map[string]int{
			"images":    0,
			"documents": 0,
			"others":    0,
		},
	}

	response.SuccessResponse(c, http.StatusOK, "Upload statistics retrieved successfully", stats)
}
