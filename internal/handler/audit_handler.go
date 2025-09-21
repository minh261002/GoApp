package handler

import (
	"net/http"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/logger"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles audit logging HTTP requests
type AuditHandler struct {
	auditService service.AuditService
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// Audit Logs

// CreateAuditLog creates a new audit log entry
// @Summary Create audit log
// @Description Create a new audit log entry
// @Tags audit
// @Accept json
// @Produce json
// @Param request body model.CreateAuditLogRequest true "Audit log creation request"
// @Success 201 {object} response.Response{data=model.AuditLogResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs [post]
func (h *AuditHandler) CreateAuditLog(c *gin.Context) {
	var req model.CreateAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Set request context
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	req.Referer = c.GetHeader("Referer")

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		userIDUint := userID.(uint)
		req.UserID = &userIDUint
	}

	log, err := h.auditService.CreateAuditLog(&req)
	if err != nil {
		logger.Errorf("Failed to create audit log: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create audit log", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Audit log created successfully", log)
}

// GetAuditLogByID retrieves an audit log by ID
// @Summary Get audit log
// @Description Get audit log by ID
// @Tags audit
// @Produce json
// @Param id path int true "Audit Log ID"
// @Success 200 {object} response.Response{data=model.AuditLogResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs/{id} [get]
func (h *AuditHandler) GetAuditLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid audit log ID", err.Error())
		return
	}

	log, err := h.auditService.GetAuditLogByID(uint(id))
	if err != nil {
		logger.Errorf("Failed to get audit log by ID %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit log", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit log retrieved successfully", log)
}

// GetAllAuditLogs retrieves all audit logs with pagination and filters
// @Summary Get all audit logs
// @Description Get all audit logs with pagination and filters
// @Tags audit
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param user_id query int false "User ID filter"
// @Param action query string false "Action filter"
// @Param resource query string false "Resource filter"
// @Param status query string false "Status filter"
// @Param severity query string false "Severity filter"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Param ip_address query string false "IP address filter"
// @Param session_id query string false "Session ID filter"
// @Param tags query string false "Tags filter (comma-separated)"
// @Success 200 {object} response.Response{data=[]model.AuditLogResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs [get]
func (h *AuditHandler) GetAllAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if resource := c.Query("resource"); resource != "" {
		filters["resource"] = resource
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}
	if ipAddress := c.Query("ip_address"); ipAddress != "" {
		filters["ip_address"] = ipAddress
	}
	if sessionID := c.Query("session_id"); sessionID != "" {
		filters["session_id"] = sessionID
	}
	if tagsStr := c.Query("tags"); tagsStr != "" {
		// Parse comma-separated tags
		tags := []string{}
		for _, tag := range []string{tagsStr} {
			tags = append(tags, tag)
		}
		filters["tags"] = tags
	}

	logs, total, err := h.auditService.GetAllAuditLogs(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all audit logs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit logs retrieved successfully", map[string]interface{}{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetAuditLogsByUser retrieves audit logs by user
// @Summary Get user audit logs
// @Description Get audit logs for a specific user
// @Tags audit
// @Produce json
// @Param user_id path int true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param action query string false "Action filter"
// @Param resource query string false "Resource filter"
// @Param status query string false "Status filter"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=[]model.AuditLogResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs/user/{user_id} [get]
func (h *AuditHandler) GetAuditLogsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if resource := c.Query("resource"); resource != "" {
		filters["resource"] = resource
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	logs, total, err := h.auditService.GetAuditLogsByUser(uint(userID), page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get audit logs by user %d: %v", userID, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user audit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User audit logs retrieved successfully", map[string]interface{}{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetAuditLogsByResource retrieves audit logs by resource
// @Summary Get resource audit logs
// @Description Get audit logs for a specific resource
// @Tags audit
// @Produce json
// @Param resource path string true "Resource type"
// @Param resource_id path int true "Resource ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=[]model.AuditLogResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs/resource/{resource}/{resource_id} [get]
func (h *AuditHandler) GetAuditLogsByResource(c *gin.Context) {
	resource := c.Param("resource")
	resourceIDStr := c.Param("resource_id")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid resource ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	logs, total, err := h.auditService.GetAuditLogsByResource(resource, uint(resourceID), page, limit)
	if err != nil {
		logger.Errorf("Failed to get audit logs by resource %s %d: %v", resource, resourceID, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve resource audit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Resource audit logs retrieved successfully", map[string]interface{}{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// SearchAuditLogs performs advanced search on audit logs
// @Summary Search audit logs
// @Description Perform advanced search on audit logs
// @Tags audit
// @Accept json
// @Produce json
// @Param request body model.AuditSearchRequest true "Search request"
// @Success 200 {object} response.Response{data=model.AuditSearchResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs/search [post]
func (h *AuditHandler) SearchAuditLogs(c *gin.Context) {
	var req model.AuditSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	searchResult, err := h.auditService.SearchAuditLogs(&req)
	if err != nil {
		logger.Errorf("Failed to search audit logs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to search audit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit logs search completed successfully", searchResult)
}

// UpdateAuditLog updates an audit log
// @Summary Update audit log
// @Description Update an existing audit log
// @Tags audit
// @Accept json
// @Produce json
// @Param id path int true "Audit Log ID"
// @Param request body model.UpdateAuditLogRequest true "Update request"
// @Success 200 {object} response.Response{data=model.AuditLogResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs/{id} [put]
func (h *AuditHandler) UpdateAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid audit log ID", err.Error())
		return
	}

	var req model.UpdateAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	log, err := h.auditService.UpdateAuditLog(uint(id), &req)
	if err != nil {
		logger.Errorf("Failed to update audit log %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update audit log", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit log updated successfully", log)
}

// DeleteAuditLog deletes an audit log
// @Summary Delete audit log
// @Description Delete an audit log
// @Tags audit
// @Param id path int true "Audit Log ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/logs/{id} [delete]
func (h *AuditHandler) DeleteAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid audit log ID", err.Error())
		return
	}

	err = h.auditService.DeleteAuditLog(uint(id))
	if err != nil {
		logger.Errorf("Failed to delete audit log %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete audit log", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit log deleted successfully", nil)
}

// Audit Config

// CreateAuditConfig creates a new audit log configuration
// @Summary Create audit config
// @Description Create a new audit log configuration
// @Tags audit
// @Accept json
// @Produce json
// @Param request body model.CreateAuditConfigRequest true "Config creation request"
// @Success 201 {object} response.Response{data=model.AuditLogConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/configs [post]
func (h *AuditHandler) CreateAuditConfig(c *gin.Context) {
	var req model.CreateAuditConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	config, err := h.auditService.CreateAuditConfig(&req)
	if err != nil {
		logger.Errorf("Failed to create audit config: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create audit config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Audit config created successfully", config)
}

// GetAuditConfigByID retrieves an audit config by ID
// @Summary Get audit config
// @Description Get audit config by ID
// @Tags audit
// @Produce json
// @Param id path int true "Config ID"
// @Success 200 {object} response.Response{data=model.AuditLogConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/configs/{id} [get]
func (h *AuditHandler) GetAuditConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID", err.Error())
		return
	}

	config, err := h.auditService.GetAuditConfigByID(uint(id))
	if err != nil {
		logger.Errorf("Failed to get audit config by ID %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit config retrieved successfully", config)
}

// GetAuditConfigByName retrieves an audit config by name
// @Summary Get audit config by name
// @Description Get audit config by name
// @Tags audit
// @Produce json
// @Param name path string true "Config name"
// @Success 200 {object} response.Response{data=model.AuditLogConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/configs/name/{name} [get]
func (h *AuditHandler) GetAuditConfigByName(c *gin.Context) {
	name := c.Param("name")

	config, err := h.auditService.GetAuditConfigByName(name)
	if err != nil {
		logger.Errorf("Failed to get audit config by name %s: %v", name, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit config retrieved successfully", config)
}

// GetAllAuditConfigs retrieves all audit configs with pagination and filters
// @Summary Get all audit configs
// @Description Get all audit configs with pagination and filters
// @Tags audit
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param is_enabled query bool false "Enabled filter"
// @Param log_level query string false "Log level filter"
// @Success 200 {object} response.Response{data=[]model.AuditLogConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/configs [get]
func (h *AuditHandler) GetAllAuditConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if isEnabledStr := c.Query("is_enabled"); isEnabledStr != "" {
		if isEnabledStr == "true" {
			filters["is_enabled"] = true
		} else if isEnabledStr == "false" {
			filters["is_enabled"] = false
		}
	}
	if logLevel := c.Query("log_level"); logLevel != "" {
		filters["log_level"] = logLevel
	}

	configs, total, err := h.auditService.GetAllAuditConfigs(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all audit configs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit configs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit configs retrieved successfully", map[string]interface{}{
		"configs": configs,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// UpdateAuditConfig updates an audit config
// @Summary Update audit config
// @Description Update an existing audit config
// @Tags audit
// @Accept json
// @Produce json
// @Param id path int true "Config ID"
// @Param request body model.UpdateAuditConfigRequest true "Update request"
// @Success 200 {object} response.Response{data=model.AuditLogConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/configs/{id} [put]
func (h *AuditHandler) UpdateAuditConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID", err.Error())
		return
	}

	var req model.UpdateAuditConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	config, err := h.auditService.UpdateAuditConfig(uint(id), &req)
	if err != nil {
		logger.Errorf("Failed to update audit config %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update audit config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit config updated successfully", config)
}

// DeleteAuditConfig deletes an audit config
// @Summary Delete audit config
// @Description Delete an audit config
// @Tags audit
// @Param id path int true "Config ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/configs/{id} [delete]
func (h *AuditHandler) DeleteAuditConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID", err.Error())
		return
	}

	err = h.auditService.DeleteAuditConfig(uint(id))
	if err != nil {
		logger.Errorf("Failed to delete audit config %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete audit config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit config deleted successfully", nil)
}

// Statistics

// GetAuditStats retrieves comprehensive audit statistics
// @Summary Get audit statistics
// @Description Get comprehensive audit statistics for the specified period
// @Tags audit
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param user_id query int false "User ID filter"
// @Param action query string false "Action filter"
// @Param resource query string false "Resource filter"
// @Success 200 {object} response.Response{data=model.AuditStats}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/stats [get]
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Missing required parameters", "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "start_date must be in YYYY-MM-DD format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "end_date must be in YYYY-MM-DD format")
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if resource := c.Query("resource"); resource != "" {
		filters["resource"] = resource
	}

	stats, err := h.auditService.GetAuditStats(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get audit stats: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit statistics retrieved successfully", stats)
}

// GetRecentActivity retrieves recent audit log activity
// @Summary Get recent activity
// @Description Get recent audit log activity
// @Tags audit
// @Produce json
// @Param limit query int false "Number of recent activities" default(20)
// @Success 200 {object} response.Response{data=[]model.AuditLogResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/activity [get]
func (h *AuditHandler) GetRecentActivity(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	activity, err := h.auditService.GetRecentActivity(limit)
	if err != nil {
		logger.Errorf("Failed to get recent activity: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve recent activity", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Recent activity retrieved successfully", activity)
}

// Export

// ExportAuditLogs exports audit logs in specified format
// @Summary Export audit logs
// @Description Export audit logs in specified format
// @Tags audit
// @Accept json
// @Produce json
// @Param request body model.AuditExportRequest true "Export request"
// @Success 200 {object} response.Response{data=model.AuditExportResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/export [post]
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	var req model.AuditExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	export, err := h.auditService.ExportAuditLogs(&req)
	if err != nil {
		logger.Errorf("Failed to export audit logs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to export audit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit logs export initiated successfully", export)
}

// Cleanup

// CleanupOldLogs cleans up old audit logs
// @Summary Cleanup old logs
// @Description Clean up old audit logs based on retention policy
// @Tags audit
// @Param retention_days query int false "Retention days" default(90)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/cleanup/logs [post]
func (h *AuditHandler) CleanupOldLogs(c *gin.Context) {
	retentionDays, _ := strconv.Atoi(c.DefaultQuery("retention_days", "90"))
	if retentionDays <= 0 {
		retentionDays = 90
	}

	err := h.auditService.CleanupOldLogs(retentionDays)
	if err != nil {
		logger.Errorf("Failed to cleanup old logs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cleanup old logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Old logs cleaned up successfully", nil)
}

// CleanupOldSummaries cleans up old audit summaries
// @Summary Cleanup old summaries
// @Description Clean up old audit summaries based on retention policy
// @Tags audit
// @Param retention_days query int false "Retention days" default(365)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/cleanup/summaries [post]
func (h *AuditHandler) CleanupOldSummaries(c *gin.Context) {
	retentionDays, _ := strconv.Atoi(c.DefaultQuery("retention_days", "365"))
	if retentionDays <= 0 {
		retentionDays = 365
	}

	err := h.auditService.CleanupOldSummaries(retentionDays)
	if err != nil {
		logger.Errorf("Failed to cleanup old summaries: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cleanup old summaries", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Old summaries cleaned up successfully", nil)
}

// OptimizeAuditTables optimizes audit log tables
// @Summary Optimize audit tables
// @Description Optimize audit log tables for better performance
// @Tags audit
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/audit/optimize [post]
func (h *AuditHandler) OptimizeAuditTables(c *gin.Context) {
	err := h.auditService.OptimizeAuditTables()
	if err != nil {
		logger.Errorf("Failed to optimize audit tables: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to optimize audit tables", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Audit tables optimized successfully", nil)
}
