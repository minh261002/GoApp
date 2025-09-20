package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/logger"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

type RateLimitHandler struct {
	rateLimitService service.RateLimitService
}

func NewRateLimitHandler(rateLimitService service.RateLimitService) *RateLimitHandler {
	return &RateLimitHandler{
		rateLimitService: rateLimitService,
	}
}

// Rate Limit Rules
// CreateRateLimitRule creates a new rate limit rule
// @Summary Create rate limit rule
// @Description Create a new rate limit rule
// @Tags rate-limit-rules
// @Accept json
// @Produce json
// @Param rule body model.RateLimitRuleRequest true "Rate limit rule data"
// @Success 201 {object} response.Response{data=model.RateLimitRuleResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/rules [post]
func (h *RateLimitHandler) CreateRateLimitRule(c *gin.Context) {
	var req model.RateLimitRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	rule, err := h.rateLimitService.CreateRateLimitRule(&req)
	if err != nil {
		logger.Errorf("Failed to create rate limit rule: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create rate limit rule", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Rate limit rule created successfully", rule)
}

// GetRateLimitRuleByID gets a rate limit rule by ID
// @Summary Get rate limit rule by ID
// @Description Get a rate limit rule by its ID
// @Tags rate-limit-rules
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} response.Response{data=model.RateLimitRuleResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/rules/{id} [get]
func (h *RateLimitHandler) GetRateLimitRuleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid rule ID", err.Error())
		return
	}

	rule, err := h.rateLimitService.GetRateLimitRuleByID(uint(id))
	if err != nil {
		if err.Error() == "rate limit rule not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Rate limit rule not found", nil)
			return
		}
		logger.Errorf("Failed to get rate limit rule: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get rate limit rule", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit rule retrieved successfully", rule)
}

// GetAllRateLimitRules gets all rate limit rules
// @Summary Get all rate limit rules
// @Description Get all rate limit rules
// @Tags rate-limit-rules
// @Produce json
// @Success 200 {object} response.Response{data=[]model.RateLimitRuleResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/rules [get]
func (h *RateLimitHandler) GetAllRateLimitRules(c *gin.Context) {
	rules, err := h.rateLimitService.GetAllRateLimitRules()
	if err != nil {
		logger.Errorf("Failed to get all rate limit rules: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get rate limit rules", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit rules retrieved successfully", rules)
}

// GetActiveRateLimitRules gets active rate limit rules
// @Summary Get active rate limit rules
// @Description Get active rate limit rules
// @Tags rate-limit-rules
// @Produce json
// @Success 200 {object} response.Response{data=[]model.RateLimitRuleResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/rules/active [get]
func (h *RateLimitHandler) GetActiveRateLimitRules(c *gin.Context) {
	rules, err := h.rateLimitService.GetActiveRateLimitRules()
	if err != nil {
		logger.Errorf("Failed to get active rate limit rules: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active rate limit rules", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Active rate limit rules retrieved successfully", rules)
}

// UpdateRateLimitRule updates a rate limit rule
// @Summary Update rate limit rule
// @Description Update an existing rate limit rule
// @Tags rate-limit-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param rule body model.RateLimitRuleRequest true "Rate limit rule data"
// @Success 200 {object} response.Response{data=model.RateLimitRuleResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/rules/{id} [put]
func (h *RateLimitHandler) UpdateRateLimitRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid rule ID", err.Error())
		return
	}

	var req model.RateLimitRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	rule, err := h.rateLimitService.UpdateRateLimitRule(uint(id), &req)
	if err != nil {
		if err.Error() == "rate limit rule not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Rate limit rule not found", nil)
			return
		}
		logger.Errorf("Failed to update rate limit rule: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update rate limit rule", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit rule updated successfully", rule)
}

// DeleteRateLimitRule deletes a rate limit rule
// @Summary Delete rate limit rule
// @Description Delete a rate limit rule by ID
// @Tags rate-limit-rules
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/rules/{id} [delete]
func (h *RateLimitHandler) DeleteRateLimitRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid rule ID", err.Error())
		return
	}

	err = h.rateLimitService.DeleteRateLimitRule(uint(id))
	if err != nil {
		if err.Error() == "rate limit rule not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Rate limit rule not found", nil)
			return
		}
		logger.Errorf("Failed to delete rate limit rule: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete rate limit rule", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit rule deleted successfully", nil)
}

// Rate Limit Logs
// GetRateLimitLogs gets rate limit logs with pagination
// @Summary Get rate limit logs
// @Description Get rate limit logs with pagination and filters
// @Tags rate-limit-logs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param rule_id query int false "Filter by rule ID"
// @Param client_ip query string false "Filter by client IP"
// @Param user_id query int false "Filter by user ID"
// @Param violation_type query string false "Filter by violation type"
// @Param is_blocked query bool false "Filter by blocked status"
// @Success 200 {object} response.Response{data=[]model.RateLimitLogResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/logs [get]
func (h *RateLimitHandler) GetRateLimitLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filters := make(map[string]interface{})
	if ruleID := c.Query("rule_id"); ruleID != "" {
		if id, err := strconv.ParseUint(ruleID, 10, 32); err == nil {
			filters["rule_id"] = id
		}
	}
	if clientIP := c.Query("client_ip"); clientIP != "" {
		filters["client_ip"] = clientIP
	}
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = id
		}
	}
	if violationType := c.Query("violation_type"); violationType != "" {
		filters["violation_type"] = violationType
	}
	if isBlocked := c.Query("is_blocked"); isBlocked != "" {
		if blocked, err := strconv.ParseBool(isBlocked); err == nil {
			filters["is_blocked"] = blocked
		}
	}

	logs, total, err := h.rateLimitService.GetRateLimitLogs(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get rate limit logs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get rate limit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit logs retrieved successfully", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetRateLimitLogsByRule gets rate limit logs by rule
// @Summary Get rate limit logs by rule
// @Description Get rate limit logs for a specific rule
// @Tags rate-limit-logs
// @Produce json
// @Param rule_id path int true "Rule ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=[]model.RateLimitLogResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/logs/rule/{rule_id} [get]
func (h *RateLimitHandler) GetRateLimitLogsByRule(c *gin.Context) {
	ruleIDStr := c.Param("rule_id")
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid rule ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	logs, total, err := h.rateLimitService.GetRateLimitLogsByRule(uint(ruleID), page, limit)
	if err != nil {
		logger.Errorf("Failed to get rate limit logs by rule: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get rate limit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit logs retrieved successfully", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetRateLimitLogsByClient gets rate limit logs by client
// @Summary Get rate limit logs by client
// @Description Get rate limit logs for a specific client IP
// @Tags rate-limit-logs
// @Produce json
// @Param client_ip path string true "Client IP"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=[]model.RateLimitLogResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/logs/client/{client_ip} [get]
func (h *RateLimitHandler) GetRateLimitLogsByClient(c *gin.Context) {
	clientIP := c.Param("client_ip")
	if clientIP == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Client IP is required", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	logs, total, err := h.rateLimitService.GetRateLimitLogsByClient(clientIP, page, limit)
	if err != nil {
		logger.Errorf("Failed to get rate limit logs by client: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get rate limit logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit logs retrieved successfully", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// CleanupOldLogs cleans up old rate limit logs
// @Summary Cleanup old rate limit logs
// @Description Cleanup rate limit logs older than specified days
// @Tags rate-limit-logs
// @Produce json
// @Param days query int false "Days to keep" default(30)
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/logs/cleanup [post]
func (h *RateLimitHandler) CleanupOldLogs(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid days parameter", err.Error())
		return
	}

	err = h.rateLimitService.CleanupOldLogs(days)
	if err != nil {
		logger.Errorf("Failed to cleanup old rate limit logs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cleanup old logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Old rate limit logs cleaned up successfully", nil)
}

// Rate Limit Stats
// GetRateLimitStats gets rate limit statistics
// @Summary Get rate limit stats
// @Description Get rate limit statistics
// @Tags rate-limit-stats
// @Produce json
// @Param rule_id query int false "Filter by rule ID"
// @Param period query string false "Filter by period" Enums(hour, day, week, month)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=[]model.RateLimitStatsResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/stats [get]
func (h *RateLimitHandler) GetRateLimitStats(c *gin.Context) {
	ruleIDStr := c.Query("rule_id")
	period := c.Query("period")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var ruleID uint
	if ruleIDStr != "" {
		if id, err := strconv.ParseUint(ruleIDStr, 10, 32); err == nil {
			ruleID = uint(id)
		}
	}

	var stats []model.RateLimitStatsResponse
	var total int64
	var err error

	if ruleID > 0 {
		stats, total, err = h.rateLimitService.GetRateLimitStats(ruleID, period, page, limit)
	} else {
		stats, total, err = h.rateLimitService.GetRateLimitStatsByPeriod(period, page, limit)
	}

	if err != nil {
		logger.Errorf("Failed to get rate limit stats: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get rate limit stats", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit stats retrieved successfully", gin.H{
		"stats": stats,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// CleanupOldStats cleans up old rate limit stats
// @Summary Cleanup old rate limit stats
// @Description Cleanup rate limit stats older than specified days
// @Tags rate-limit-stats
// @Produce json
// @Param days query int false "Days to keep" default(90)
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/stats/cleanup [post]
func (h *RateLimitHandler) CleanupOldStats(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "90")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid days parameter", err.Error())
		return
	}

	err = h.rateLimitService.CleanupOldStats(days)
	if err != nil {
		logger.Errorf("Failed to cleanup old rate limit stats: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cleanup old stats", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Old rate limit stats cleaned up successfully", nil)
}

// Whitelist/Blacklist
// CreateWhitelistEntry creates a whitelist entry
// @Summary Create whitelist entry
// @Description Create a new whitelist entry
// @Tags rate-limit-whitelist
// @Accept json
// @Produce json
// @Param entry body model.RateLimitWhitelistRequest true "Whitelist entry data"
// @Success 201 {object} response.Response{data=model.RateLimitWhitelist}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/whitelist [post]
func (h *RateLimitHandler) CreateWhitelistEntry(c *gin.Context) {
	var req model.RateLimitWhitelistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	entry, err := h.rateLimitService.CreateWhitelistEntry(&req)
	if err != nil {
		logger.Errorf("Failed to create whitelist entry: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create whitelist entry", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Whitelist entry created successfully", entry)
}

// GetWhitelistEntries gets all whitelist entries
// @Summary Get whitelist entries
// @Description Get all whitelist entries
// @Tags rate-limit-whitelist
// @Produce json
// @Success 200 {object} response.Response{data=[]model.RateLimitWhitelist}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/whitelist [get]
func (h *RateLimitHandler) GetWhitelistEntries(c *gin.Context) {
	entries, err := h.rateLimitService.GetAllWhitelistEntries()
	if err != nil {
		logger.Errorf("Failed to get whitelist entries: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get whitelist entries", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Whitelist entries retrieved successfully", entries)
}

// GetActiveWhitelistEntries gets active whitelist entries
// @Summary Get active whitelist entries
// @Description Get active whitelist entries
// @Tags rate-limit-whitelist
// @Produce json
// @Success 200 {object} response.Response{data=[]model.RateLimitWhitelist}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/whitelist/active [get]
func (h *RateLimitHandler) GetActiveWhitelistEntries(c *gin.Context) {
	entries, err := h.rateLimitService.GetActiveWhitelistEntries()
	if err != nil {
		logger.Errorf("Failed to get active whitelist entries: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active whitelist entries", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Active whitelist entries retrieved successfully", entries)
}

// DeleteWhitelistEntry deletes a whitelist entry
// @Summary Delete whitelist entry
// @Description Delete a whitelist entry by ID
// @Tags rate-limit-whitelist
// @Produce json
// @Param id path int true "Entry ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/whitelist/{id} [delete]
func (h *RateLimitHandler) DeleteWhitelistEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid entry ID", err.Error())
		return
	}

	err = h.rateLimitService.DeleteWhitelistEntry(uint(id))
	if err != nil {
		if err.Error() == "whitelist entry not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Whitelist entry not found", nil)
			return
		}
		logger.Errorf("Failed to delete whitelist entry: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete whitelist entry", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Whitelist entry deleted successfully", nil)
}

// Blacklist endpoints (similar structure to whitelist)
// CreateBlacklistEntry creates a blacklist entry
// @Summary Create blacklist entry
// @Description Create a new blacklist entry
// @Tags rate-limit-blacklist
// @Accept json
// @Produce json
// @Param entry body model.RateLimitBlacklistRequest true "Blacklist entry data"
// @Success 201 {object} response.Response{data=model.RateLimitBlacklist}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/blacklist [post]
func (h *RateLimitHandler) CreateBlacklistEntry(c *gin.Context) {
	var req model.RateLimitBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	entry, err := h.rateLimitService.CreateBlacklistEntry(&req)
	if err != nil {
		logger.Errorf("Failed to create blacklist entry: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create blacklist entry", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Blacklist entry created successfully", entry)
}

// GetBlacklistEntries gets all blacklist entries
// @Summary Get blacklist entries
// @Description Get all blacklist entries
// @Tags rate-limit-blacklist
// @Produce json
// @Success 200 {object} response.Response{data=[]model.RateLimitBlacklist}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/blacklist [get]
func (h *RateLimitHandler) GetBlacklistEntries(c *gin.Context) {
	entries, err := h.rateLimitService.GetAllBlacklistEntries()
	if err != nil {
		logger.Errorf("Failed to get blacklist entries: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get blacklist entries", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Blacklist entries retrieved successfully", entries)
}

// GetActiveBlacklistEntries gets active blacklist entries
// @Summary Get active blacklist entries
// @Description Get active blacklist entries
// @Tags rate-limit-blacklist
// @Produce json
// @Success 200 {object} response.Response{data=[]model.RateLimitBlacklist}
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/blacklist/active [get]
func (h *RateLimitHandler) GetActiveBlacklistEntries(c *gin.Context) {
	entries, err := h.rateLimitService.GetActiveBlacklistEntries()
	if err != nil {
		logger.Errorf("Failed to get active blacklist entries: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active blacklist entries", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Active blacklist entries retrieved successfully", entries)
}

// DeleteBlacklistEntry deletes a blacklist entry
// @Summary Delete blacklist entry
// @Description Delete a blacklist entry by ID
// @Tags rate-limit-blacklist
// @Produce json
// @Param id path int true "Entry ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/blacklist/{id} [delete]
func (h *RateLimitHandler) DeleteBlacklistEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid entry ID", err.Error())
		return
	}

	err = h.rateLimitService.DeleteBlacklistEntry(uint(id))
	if err != nil {
		if err.Error() == "blacklist entry not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Blacklist entry not found", nil)
			return
		}
		logger.Errorf("Failed to delete blacklist entry: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete blacklist entry", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Blacklist entry deleted successfully", nil)
}

// Rate Limit Info
// GetRateLimitInfo gets current rate limit info for a client
// @Summary Get rate limit info
// @Description Get current rate limit information for a client
// @Tags rate-limit-info
// @Produce json
// @Param rule_name query string true "Rule name"
// @Param client_id query string true "Client ID"
// @Success 200 {object} response.Response{data=ratelimit.RateLimitInfo}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/info [get]
func (h *RateLimitHandler) GetRateLimitInfo(c *gin.Context) {
	ruleName := c.Query("rule_name")
	clientID := c.Query("client_id")

	if ruleName == "" || clientID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Rule name and client ID are required", nil)
		return
	}

	info, err := h.rateLimitService.GetRateLimitInfo(ruleName, clientID)
	if err != nil {
		logger.Errorf("Failed to get rate limit info: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get rate limit info", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit info retrieved successfully", info)
}

// ClearRateLimit clears rate limit for a client
// @Summary Clear rate limit
// @Description Clear rate limit for a specific client
// @Tags rate-limit-info
// @Produce json
// @Param rule_name query string true "Rule name"
// @Param client_id query string true "Client ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rate-limit/clear [post]
func (h *RateLimitHandler) ClearRateLimit(c *gin.Context) {
	ruleName := c.Query("rule_name")
	clientID := c.Query("client_id")

	if ruleName == "" || clientID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Rule name and client ID are required", nil)
		return
	}

	err := h.rateLimitService.ClearRateLimit(ruleName, clientID)
	if err != nil {
		logger.Errorf("Failed to clear rate limit: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to clear rate limit", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rate limit cleared successfully", nil)
}
