package middleware

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"go_app/configs"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// UploadMiddleware handles file upload validation and size limits
type UploadMiddleware struct {
	config *configs.Config
}

// NewUploadMiddleware creates a new upload middleware
func NewUploadMiddleware() *UploadMiddleware {
	return &UploadMiddleware{
		config: configs.Load(),
	}
}

// FileSizeLimit limits the file size
func (m *UploadMiddleware) FileSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set the max memory for multipart forms
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// ValidateFileExtension validates file extensions
func (m *UploadMiddleware) ValidateFileExtension(allowedExts []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the file from form data
		file, err := c.FormFile("file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
			c.Abort()
			return
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == "" {
			ext = ".bin"
		}

		// Check if extension is allowed
		allowed := false
		for _, allowedExt := range allowedExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Error(c, http.StatusBadRequest, "File extension not allowed",
				fmt.Sprintf("File extension %s is not allowed. Allowed extensions: %s", ext, strings.Join(allowedExts, ", ")))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateImageFile validates image files specifically
func (m *UploadMiddleware) ValidateImageFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the file from form data
		file, err := c.FormFile("file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
			c.Abort()
			return
		}

		// Check file size (max 5MB for images)
		if file.Size > m.config.Upload.ImageMaxSize {
			response.Error(c, http.StatusRequestEntityTooLarge, "File too large",
				fmt.Sprintf("Image file size exceeds maximum allowed size of %d bytes", m.config.Upload.ImageMaxSize))
			c.Abort()
			return
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
			response.Error(c, http.StatusBadRequest, "Invalid image file",
				fmt.Sprintf("Image file extension %s is not allowed. Allowed extensions: %s", ext, strings.Join(allowedImageExts, ", ")))
			c.Abort()
			return
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
			response.Error(c, http.StatusBadRequest, "Invalid image file",
				fmt.Sprintf("Image MIME type %s is not allowed. Allowed types: %s", mimeType, strings.Join(allowedMimeTypes, ", ")))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateDocumentFile validates document files
func (m *UploadMiddleware) ValidateDocumentFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the file from form data
		file, err := c.FormFile("file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
			c.Abort()
			return
		}

		// Check file size (max 20MB for documents)
		if file.Size > m.config.Upload.DocumentMaxSize {
			response.Error(c, http.StatusRequestEntityTooLarge, "File too large",
				fmt.Sprintf("Document file size exceeds maximum allowed size of %d bytes", m.config.Upload.DocumentMaxSize))
			c.Abort()
			return
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowedDocExts := []string{".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx", ".xls"}

		allowed := false
		for _, allowedExt := range allowedDocExts {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			response.Error(c, http.StatusBadRequest, "Invalid document file",
				fmt.Sprintf("Document file extension %s is not allowed. Allowed extensions: %s", ext, strings.Join(allowedDocExts, ", ")))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateMultipleFiles validates multiple files
func (m *UploadMiddleware) ValidateMultipleFiles(maxFiles int, allowedExts []string, maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Failed to parse form data", err.Error())
			c.Abort()
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			response.Error(c, http.StatusBadRequest, "No files provided", "At least one file is required")
			c.Abort()
			return
		}

		if len(files) > maxFiles {
			response.Error(c, http.StatusBadRequest, "Too many files",
				fmt.Sprintf("Maximum %d files allowed", maxFiles))
			c.Abort()
			return
		}

		// Validate each file
		for i, file := range files {
			// Check file size
			if file.Size > maxSize {
				response.Error(c, http.StatusRequestEntityTooLarge, "File too large",
					fmt.Sprintf("File %d (%s) size exceeds maximum allowed size of %d bytes", i+1, file.Filename, maxSize))
				c.Abort()
				return
			}

			// Check file extension
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if ext == "" {
				ext = ".bin"
			}

			allowed := false
			for _, allowedExt := range allowedExts {
				if ext == strings.ToLower(allowedExt) {
					allowed = true
					break
				}
			}
			if !allowed {
				response.Error(c, http.StatusBadRequest, "Invalid file extension",
					fmt.Sprintf("File %d (%s) extension %s is not allowed. Allowed extensions: %s", i+1, file.Filename, ext, strings.Join(allowedExts, ", ")))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// SecurityScan performs basic security checks on uploaded files
func (m *UploadMiddleware) SecurityScan() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the file from form data
		file, err := c.FormFile("file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "No file provided", err.Error())
			c.Abort()
			return
		}

		// Open file for reading
		src, err := file.Open()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to open file", err.Error())
			c.Abort()
			return
		}
		defer src.Close()

		// Read first 512 bytes for magic number detection
		buffer := make([]byte, 512)
		_, err = src.Read(buffer)
		if err != nil && err != io.EOF {
			response.Error(c, http.StatusInternalServerError, "Failed to read file", err.Error())
			c.Abort()
			return
		}

		// Check for suspicious content
		if m.containsSuspiciousContent(buffer) {
			response.Error(c, http.StatusBadRequest, "File rejected", "File contains suspicious content")
			c.Abort()
			return
		}

		c.Next()
	}
}

// containsSuspiciousContent checks for suspicious content in file header
func (m *UploadMiddleware) containsSuspiciousContent(buffer []byte) bool {
	// Check for executable signatures
	executableSignatures := [][]byte{
		{0x4D, 0x5A},             // PE executable
		{0x7F, 0x45, 0x4C, 0x46}, // ELF executable
		{0xCA, 0xFE, 0xBA, 0xBE}, // Java class file
		{0xFE, 0xED, 0xFA, 0xCE}, // Mach-O executable
	}

	for _, sig := range executableSignatures {
		if len(buffer) >= len(sig) {
			match := true
			for i, b := range sig {
				if buffer[i] != b {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	// Check for script signatures
	scriptSignatures := [][]byte{
		{0x3C, 0x3F, 0x78, 0x6D, 0x6C}, // XML
		{0x3C, 0x68, 0x74, 0x6D, 0x6C}, // HTML
		{0x3C, 0x21, 0x44, 0x4F, 0x43}, // HTML DOCTYPE
	}

	for _, sig := range scriptSignatures {
		if len(buffer) >= len(sig) {
			match := true
			for i, b := range sig {
				if buffer[i] != b {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	return false
}

// RateLimitUpload limits upload rate per IP
func (m *UploadMiddleware) RateLimitUpload(maxUploadsPerMinute int) gin.HandlerFunc {
	// This is a simple implementation
	// In production, you should use Redis or similar for rate limiting
	return func(c *gin.Context) {
		// For now, just pass through
		// TODO: Implement proper rate limiting
		c.Next()
	}
}
