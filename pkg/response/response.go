package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response cấu trúc response chuẩn
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta thông tin phân trang và metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// PaginationRequest cấu trúc request phân trang
type PaginationRequest struct {
	Page  int `form:"page" binding:"min=1"`
	Limit int `form:"limit" binding:"min=1,max=100"`
}

// Success trả về response thành công
func Success(c *gin.Context, data interface{}, message ...string) {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

// SuccessWithMeta trả về response thành công với metadata
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta, message ...string) {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: msg,
		Data:    data,
		Meta:    meta,
	})
}

// Created trả về response tạo mới thành công
func Created(c *gin.Context, data interface{}, message ...string) {
	msg := "Created successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

// BadRequest trả về response lỗi 400
func BadRequest(c *gin.Context, message string, err ...interface{}) {
	response := Response{
		Success: false,
		Message: message,
	}

	if len(err) > 0 {
		response.Error = err[0]
	}

	c.JSON(http.StatusBadRequest, response)
}

// Unauthorized trả về response lỗi 401
func Unauthorized(c *gin.Context, message ...string) {
	msg := "Unauthorized"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Message: msg,
	})
}

// Forbidden trả về response lỗi 403
func Forbidden(c *gin.Context, message ...string) {
	msg := "Forbidden"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Message: msg,
	})
}

// NotFound trả về response lỗi 404
func NotFound(c *gin.Context, message ...string) {
	msg := "Not found"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: msg,
	})
}

// Conflict trả về response lỗi 409
func Conflict(c *gin.Context, message ...string) {
	msg := "Conflict"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusConflict, Response{
		Success: false,
		Message: msg,
	})
}

// ValidationError trả về response lỗi validation
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Message: "Validation failed",
		Error:   errors,
	})
}

// InternalServerError trả về response lỗi 500
func InternalServerError(c *gin.Context, message ...string) {
	msg := "Internal server error"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: msg,
	})
}

// Error trả về response lỗi tùy chỉnh
func Error(c *gin.Context, statusCode int, message string, err ...interface{}) {
	response := Response{
		Success: false,
		Message: message,
	}

	if len(err) > 0 {
		response.Error = err[0]
	}

	c.JSON(statusCode, response)
}

// PaginationResponse tạo response phân trang
func PaginationResponse(c *gin.Context, data interface{}, page, limit int, total int64) {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	meta := &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	SuccessWithMeta(c, data, meta)
}

// HealthCheck trả về response health check
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Service is running",
	})
}

// ErrorResponse trả về response lỗi với status code tùy chỉnh
func ErrorResponse(c *gin.Context, statusCode int, message string, err ...interface{}) {
	Error(c, statusCode, message, err...)
}

// SuccessResponse trả về response thành công với status code tùy chỉnh
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	msg := "Success"
	if message != "" {
		msg = message
	}

	c.JSON(statusCode, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

// SuccessResponseWithPagination trả về response thành công với phân trang
func SuccessResponseWithPagination(c *gin.Context, statusCode int, message string, data interface{}, page, limit int, total int64) {
	msg := "Success"
	if message != "" {
		msg = message
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	meta := &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(statusCode, Response{
		Success: true,
		Message: msg,
		Data:    data,
		Meta:    meta,
	})
}
