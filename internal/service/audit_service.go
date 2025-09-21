package service

import (
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"time"
)

// AuditService defines methods for audit logging business logic
type AuditService interface {
	// Audit Logs
	CreateAuditLog(req *model.CreateAuditLogRequest) (*model.AuditLogResponse, error)
	GetAuditLogByID(id uint) (*model.AuditLogResponse, error)
	GetAllAuditLogs(page, limit int, filters map[string]interface{}) ([]model.AuditLogResponse, int64, error)
	GetAuditLogsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AuditLogResponse, int64, error)
	GetAuditLogsByResource(resource string, resourceID uint, page, limit int) ([]model.AuditLogResponse, int64, error)
	GetAuditLogsByAction(action string, page, limit int, filters map[string]interface{}) ([]model.AuditLogResponse, int64, error)
	SearchAuditLogs(req *model.AuditSearchRequest) (*model.AuditSearchResponse, error)
	UpdateAuditLog(id uint, req *model.UpdateAuditLogRequest) (*model.AuditLogResponse, error)
	DeleteAuditLog(id uint) error

	// Audit Config
	CreateAuditConfig(req *model.CreateAuditConfigRequest) (*model.AuditLogConfigResponse, error)
	GetAuditConfigByID(id uint) (*model.AuditLogConfigResponse, error)
	GetAuditConfigByName(name string) (*model.AuditLogConfigResponse, error)
	GetAllAuditConfigs(page, limit int, filters map[string]interface{}) ([]model.AuditLogConfigResponse, int64, error)
	UpdateAuditConfig(id uint, req *model.UpdateAuditConfigRequest) (*model.AuditLogConfigResponse, error)
	DeleteAuditConfig(id uint) error

	// Audit Summaries
	GetAuditSummaries(startDate, endDate time.Time, filters map[string]interface{}) ([]model.AuditLogSummaryResponse, error)
	GenerateDailySummaries(date time.Time) error

	// Statistics
	GetAuditStats(startDate, endDate time.Time, filters map[string]interface{}) (*model.AuditStats, error)
	GetTopActions(startDate, endDate time.Time, limit int) ([]model.ActionStats, error)
	GetTopResources(startDate, endDate time.Time, limit int) ([]model.ResourceStats, error)
	GetTopUsers(startDate, endDate time.Time, limit int) ([]model.UserStats, error)
	GetDailyStats(startDate, endDate time.Time) ([]model.DailyStats, error)
	GetHourlyStats(startDate, endDate time.Time) ([]model.HourlyStats, error)
	GetRecentActivity(limit int) ([]model.AuditLogResponse, error)

	// Export
	ExportAuditLogs(req *model.AuditExportRequest) (*model.AuditExportResponse, error)

	// Cleanup
	CleanupOldLogs(retentionDays int) error
	CleanupOldSummaries(retentionDays int) error
	OptimizeAuditTables() error

	// Helper methods for logging
	LogAction(userID *uint, action, resource string, resourceID *uint, resourceName, message string, oldValues, newValues, changes, metadata map[string]interface{}, tags []string, severity string, targetUserID *uint, ipAddress, userAgent, referer, sessionID string) error
	LogUserAction(userID uint, action, resource string, resourceID *uint, resourceName, message string, oldValues, newValues, changes, metadata map[string]interface{}, tags []string, severity string, targetUserID *uint, ipAddress, userAgent, referer, sessionID string) error
	LogSystemAction(action, resource string, resourceID *uint, resourceName, message string, oldValues, newValues, changes, metadata map[string]interface{}, tags []string, severity string, ipAddress, userAgent, referer, sessionID string) error
}

// auditService implements AuditService
type auditService struct {
	auditRepo repository.AuditRepository
	userRepo  repository.UserRepository
}

// NewAuditService creates a new AuditService
func NewAuditService(auditRepo repository.AuditRepository, userRepo repository.UserRepository) AuditService {
	return &auditService{
		auditRepo: auditRepo,
		userRepo:  userRepo,
	}
}

// Audit Logs

// CreateAuditLog creates a new audit log entry
func (s *auditService) CreateAuditLog(req *model.CreateAuditLogRequest) (*model.AuditLogResponse, error) {
	// Validate required fields
	if req.Action == "" {
		return nil, fmt.Errorf("action is required")
	}
	if req.Resource == "" {
		return nil, fmt.Errorf("resource is required")
	}
	if req.Operation == "" {
		return nil, fmt.Errorf("operation is required")
	}
	if req.Status == "" {
		return nil, fmt.Errorf("status is required")
	}

	// Set default severity if not provided
	if req.Severity == "" {
		req.Severity = model.SeverityInfo
	}

	// Create audit log
	log := &model.AuditLog{
		UserID:       req.UserID,
		Action:       req.Action,
		Resource:     req.Resource,
		ResourceID:   req.ResourceID,
		ResourceName: req.ResourceName,
		Operation:    req.Operation,
		Status:       req.Status,
		Message:      req.Message,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		Referer:      req.Referer,
		SessionID:    req.SessionID,
		OldValues:    fmt.Sprintf("%v", req.OldValues), // Simplified JSON conversion
		NewValues:    fmt.Sprintf("%v", req.NewValues), // Simplified JSON conversion
		Changes:      fmt.Sprintf("%v", req.Changes),   // Simplified JSON conversion
		Metadata:     fmt.Sprintf("%v", req.Metadata),  // Simplified JSON conversion
		Tags:         fmt.Sprintf("%v", req.Tags),      // Simplified JSON conversion
		Severity:     req.Severity,
		TargetUserID: req.TargetUserID,
	}

	if err := s.auditRepo.CreateAuditLog(log); err != nil {
		logger.Errorf("Failed to create audit log: %v", err)
		return nil, fmt.Errorf("failed to create audit log")
	}

	// Get created log with relations
	createdLog, err := s.auditRepo.GetAuditLogByID(log.ID)
	if err != nil {
		logger.Errorf("Failed to get created audit log: %v", err)
		return nil, fmt.Errorf("failed to retrieve created audit log")
	}

	return s.toAuditLogResponse(createdLog), nil
}

// GetAuditLogByID retrieves an audit log by ID
func (s *auditService) GetAuditLogByID(id uint) (*model.AuditLogResponse, error) {
	log, err := s.auditRepo.GetAuditLogByID(id)
	if err != nil {
		logger.Errorf("Failed to get audit log by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve audit log")
	}
	if log == nil {
		return nil, fmt.Errorf("audit log not found")
	}

	return s.toAuditLogResponse(log), nil
}

// GetAllAuditLogs retrieves all audit logs with pagination and filters
func (s *auditService) GetAllAuditLogs(page, limit int, filters map[string]interface{}) ([]model.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetAllAuditLogs(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all audit logs: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve audit logs")
	}

	var responses []model.AuditLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toAuditLogResponse(&log))
	}

	return responses, total, nil
}

// GetAuditLogsByUser retrieves audit logs by user
func (s *auditService) GetAuditLogsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetAuditLogsByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get audit logs by user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve user audit logs")
	}

	var responses []model.AuditLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toAuditLogResponse(&log))
	}

	return responses, total, nil
}

// GetAuditLogsByResource retrieves audit logs by resource
func (s *auditService) GetAuditLogsByResource(resource string, resourceID uint, page, limit int) ([]model.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetAuditLogsByResource(resource, resourceID, page, limit)
	if err != nil {
		logger.Errorf("Failed to get audit logs by resource %s %d: %v", resource, resourceID, err)
		return nil, 0, fmt.Errorf("failed to retrieve resource audit logs")
	}

	var responses []model.AuditLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toAuditLogResponse(&log))
	}

	return responses, total, nil
}

// GetAuditLogsByAction retrieves audit logs by action
func (s *auditService) GetAuditLogsByAction(action string, page, limit int, filters map[string]interface{}) ([]model.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetAuditLogsByAction(action, page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get audit logs by action %s: %v", action, err)
		return nil, 0, fmt.Errorf("failed to retrieve action audit logs")
	}

	var responses []model.AuditLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toAuditLogResponse(&log))
	}

	return responses, total, nil
}

// SearchAuditLogs performs advanced search on audit logs
func (s *auditService) SearchAuditLogs(req *model.AuditSearchRequest) (*model.AuditSearchResponse, error) {
	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	logs, total, err := s.auditRepo.SearchAuditLogs(req)
	if err != nil {
		logger.Errorf("Failed to search audit logs: %v", err)
		return nil, fmt.Errorf("failed to search audit logs")
	}

	var responses []model.AuditLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toAuditLogResponse(&log))
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &model.AuditSearchResponse{
		Logs:       responses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}, nil
}

// UpdateAuditLog updates an audit log
func (s *auditService) UpdateAuditLog(id uint, req *model.UpdateAuditLogRequest) (*model.AuditLogResponse, error) {
	log, err := s.auditRepo.GetAuditLogByID(id)
	if err != nil {
		logger.Errorf("Failed to get audit log by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve audit log")
	}
	if log == nil {
		return nil, fmt.Errorf("audit log not found")
	}

	// Update fields
	if req.Message != "" {
		log.Message = req.Message
	}
	if req.Status != "" {
		log.Status = req.Status
	}
	if req.Severity != "" {
		log.Severity = req.Severity
	}
	if req.Metadata != nil {
		log.Metadata = fmt.Sprintf("%v", req.Metadata)
	}
	if req.Tags != nil {
		log.Tags = fmt.Sprintf("%v", req.Tags)
	}

	if err := s.auditRepo.UpdateAuditLog(log); err != nil {
		logger.Errorf("Failed to update audit log %d: %v", id, err)
		return nil, fmt.Errorf("failed to update audit log")
	}

	return s.toAuditLogResponse(log), nil
}

// DeleteAuditLog deletes an audit log
func (s *auditService) DeleteAuditLog(id uint) error {
	if err := s.auditRepo.DeleteAuditLog(id); err != nil {
		logger.Errorf("Failed to delete audit log %d: %v", id, err)
		return fmt.Errorf("failed to delete audit log")
	}
	return nil
}

// Audit Config

// CreateAuditConfig creates a new audit log configuration
func (s *auditService) CreateAuditConfig(req *model.CreateAuditConfigRequest) (*model.AuditLogConfigResponse, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Check if config with same name already exists
	existingConfig, err := s.auditRepo.GetAuditConfigByName(req.Name)
	if err != nil {
		logger.Errorf("Failed to check existing audit config: %v", err)
		return nil, fmt.Errorf("failed to check existing audit config")
	}
	if existingConfig != nil {
		return nil, fmt.Errorf("audit config with name '%s' already exists", req.Name)
	}

	// Set default values
	if req.LogLevel == "" {
		req.LogLevel = model.LogLevelInfo
	}
	if req.RetentionDays <= 0 {
		req.RetentionDays = 90
	}
	if req.MaxLogSize <= 0 {
		req.MaxLogSize = 1000000
	}

	// Create audit config
	config := &model.AuditLogConfig{
		Name:          req.Name,
		Description:   req.Description,
		IsEnabled:     req.IsEnabled,
		LogLevel:      req.LogLevel,
		Resources:     fmt.Sprintf("%v", req.Resources),    // Simplified JSON conversion
		Actions:       fmt.Sprintf("%v", req.Actions),      // Simplified JSON conversion
		ExcludeUsers:  fmt.Sprintf("%v", req.ExcludeUsers), // Simplified JSON conversion
		RetentionDays: req.RetentionDays,
		MaxLogSize:    req.MaxLogSize,
	}

	if err := s.auditRepo.CreateAuditConfig(config); err != nil {
		logger.Errorf("Failed to create audit config: %v", err)
		return nil, fmt.Errorf("failed to create audit config")
	}

	// Get created config
	createdConfig, err := s.auditRepo.GetAuditConfigByID(config.ID)
	if err != nil {
		logger.Errorf("Failed to get created audit config: %v", err)
		return nil, fmt.Errorf("failed to retrieve created audit config")
	}

	return s.toAuditConfigResponse(createdConfig), nil
}

// GetAuditConfigByID retrieves an audit config by ID
func (s *auditService) GetAuditConfigByID(id uint) (*model.AuditLogConfigResponse, error) {
	config, err := s.auditRepo.GetAuditConfigByID(id)
	if err != nil {
		logger.Errorf("Failed to get audit config by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve audit config")
	}
	if config == nil {
		return nil, fmt.Errorf("audit config not found")
	}

	return s.toAuditConfigResponse(config), nil
}

// GetAuditConfigByName retrieves an audit config by name
func (s *auditService) GetAuditConfigByName(name string) (*model.AuditLogConfigResponse, error) {
	config, err := s.auditRepo.GetAuditConfigByName(name)
	if err != nil {
		logger.Errorf("Failed to get audit config by name %s: %v", name, err)
		return nil, fmt.Errorf("failed to retrieve audit config")
	}
	if config == nil {
		return nil, fmt.Errorf("audit config not found")
	}

	return s.toAuditConfigResponse(config), nil
}

// GetAllAuditConfigs retrieves all audit configs with pagination and filters
func (s *auditService) GetAllAuditConfigs(page, limit int, filters map[string]interface{}) ([]model.AuditLogConfigResponse, int64, error) {
	configs, total, err := s.auditRepo.GetAllAuditConfigs(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all audit configs: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve audit configs")
	}

	var responses []model.AuditLogConfigResponse
	for _, config := range configs {
		responses = append(responses, *s.toAuditConfigResponse(&config))
	}

	return responses, total, nil
}

// UpdateAuditConfig updates an audit config
func (s *auditService) UpdateAuditConfig(id uint, req *model.UpdateAuditConfigRequest) (*model.AuditLogConfigResponse, error) {
	config, err := s.auditRepo.GetAuditConfigByID(id)
	if err != nil {
		logger.Errorf("Failed to get audit config by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve audit config")
	}
	if config == nil {
		return nil, fmt.Errorf("audit config not found")
	}

	// Update fields
	if req.Name != "" {
		config.Name = req.Name
	}
	if req.Description != "" {
		config.Description = req.Description
	}
	if req.IsEnabled != nil {
		config.IsEnabled = *req.IsEnabled
	}
	if req.LogLevel != "" {
		config.LogLevel = req.LogLevel
	}
	if req.Resources != nil {
		config.Resources = fmt.Sprintf("%v", req.Resources)
	}
	if req.Actions != nil {
		config.Actions = fmt.Sprintf("%v", req.Actions)
	}
	if req.ExcludeUsers != nil {
		config.ExcludeUsers = fmt.Sprintf("%v", req.ExcludeUsers)
	}
	if req.RetentionDays > 0 {
		config.RetentionDays = req.RetentionDays
	}
	if req.MaxLogSize > 0 {
		config.MaxLogSize = req.MaxLogSize
	}

	if err := s.auditRepo.UpdateAuditConfig(config); err != nil {
		logger.Errorf("Failed to update audit config %d: %v", id, err)
		return nil, fmt.Errorf("failed to update audit config")
	}

	return s.toAuditConfigResponse(config), nil
}

// DeleteAuditConfig deletes an audit config
func (s *auditService) DeleteAuditConfig(id uint) error {
	if err := s.auditRepo.DeleteAuditConfig(id); err != nil {
		logger.Errorf("Failed to delete audit config %d: %v", id, err)
		return fmt.Errorf("failed to delete audit config")
	}
	return nil
}

// Audit Summaries

// GetAuditSummaries retrieves audit summaries for a date range
func (s *auditService) GetAuditSummaries(startDate, endDate time.Time, filters map[string]interface{}) ([]model.AuditLogSummaryResponse, error) {
	summaries, err := s.auditRepo.GetAuditSummaries(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get audit summaries: %v", err)
		return nil, fmt.Errorf("failed to retrieve audit summaries")
	}

	var responses []model.AuditLogSummaryResponse
	for _, summary := range summaries {
		responses = append(responses, *s.toAuditSummaryResponse(&summary))
	}

	return responses, nil
}

// GenerateDailySummaries generates daily summaries for a specific date
func (s *auditService) GenerateDailySummaries(date time.Time) error {
	if err := s.auditRepo.GenerateDailySummaries(date); err != nil {
		logger.Errorf("Failed to generate daily summaries for %s: %v", date.Format("2006-01-02"), err)
		return fmt.Errorf("failed to generate daily summaries")
	}
	return nil
}

// Statistics

// GetAuditStats retrieves comprehensive audit statistics
func (s *auditService) GetAuditStats(startDate, endDate time.Time, filters map[string]interface{}) (*model.AuditStats, error) {
	stats, err := s.auditRepo.GetAuditStats(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get audit stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve audit statistics")
	}
	return stats, nil
}

// GetTopActions retrieves top actions by count
func (s *auditService) GetTopActions(startDate, endDate time.Time, limit int) ([]model.ActionStats, error) {
	actions, err := s.auditRepo.GetTopActions(startDate, endDate, limit)
	if err != nil {
		logger.Errorf("Failed to get top actions: %v", err)
		return nil, fmt.Errorf("failed to retrieve top actions")
	}
	return actions, nil
}

// GetTopResources retrieves top resources by count
func (s *auditService) GetTopResources(startDate, endDate time.Time, limit int) ([]model.ResourceStats, error) {
	resources, err := s.auditRepo.GetTopResources(startDate, endDate, limit)
	if err != nil {
		logger.Errorf("Failed to get top resources: %v", err)
		return nil, fmt.Errorf("failed to retrieve top resources")
	}
	return resources, nil
}

// GetTopUsers retrieves top users by activity count
func (s *auditService) GetTopUsers(startDate, endDate time.Time, limit int) ([]model.UserStats, error) {
	users, err := s.auditRepo.GetTopUsers(startDate, endDate, limit)
	if err != nil {
		logger.Errorf("Failed to get top users: %v", err)
		return nil, fmt.Errorf("failed to retrieve top users")
	}
	return users, nil
}

// GetDailyStats retrieves daily statistics
func (s *auditService) GetDailyStats(startDate, endDate time.Time) ([]model.DailyStats, error) {
	stats, err := s.auditRepo.GetDailyStats(startDate, endDate)
	if err != nil {
		logger.Errorf("Failed to get daily stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve daily statistics")
	}
	return stats, nil
}

// GetHourlyStats retrieves hourly statistics
func (s *auditService) GetHourlyStats(startDate, endDate time.Time) ([]model.HourlyStats, error) {
	stats, err := s.auditRepo.GetHourlyStats(startDate, endDate)
	if err != nil {
		logger.Errorf("Failed to get hourly stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve hourly statistics")
	}
	return stats, nil
}

// GetRecentActivity retrieves recent audit log activity
func (s *auditService) GetRecentActivity(limit int) ([]model.AuditLogResponse, error) {
	logs, err := s.auditRepo.GetRecentActivity(limit)
	if err != nil {
		logger.Errorf("Failed to get recent activity: %v", err)
		return nil, fmt.Errorf("failed to retrieve recent activity")
	}

	var responses []model.AuditLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toAuditLogResponse(&log))
	}

	return responses, nil
}

// Export

// ExportAuditLogs exports audit logs in specified format
func (s *auditService) ExportAuditLogs(req *model.AuditExportRequest) (*model.AuditExportResponse, error) {
	response, err := s.auditRepo.ExportAuditLogs(req)
	if err != nil {
		logger.Errorf("Failed to export audit logs: %v", err)
		return nil, fmt.Errorf("failed to export audit logs")
	}
	return response, nil
}

// Cleanup

// CleanupOldLogs cleans up old audit logs
func (s *auditService) CleanupOldLogs(retentionDays int) error {
	if err := s.auditRepo.CleanupOldLogs(retentionDays); err != nil {
		logger.Errorf("Failed to cleanup old logs: %v", err)
		return fmt.Errorf("failed to cleanup old logs")
	}
	return nil
}

// CleanupOldSummaries cleans up old audit summaries
func (s *auditService) CleanupOldSummaries(retentionDays int) error {
	if err := s.auditRepo.CleanupOldSummaries(retentionDays); err != nil {
		logger.Errorf("Failed to cleanup old summaries: %v", err)
		return fmt.Errorf("failed to cleanup old summaries")
	}
	return nil
}

// OptimizeAuditTables optimizes audit log tables
func (s *auditService) OptimizeAuditTables() error {
	if err := s.auditRepo.OptimizeAuditTables(); err != nil {
		logger.Errorf("Failed to optimize audit tables: %v", err)
		return fmt.Errorf("failed to optimize audit tables")
	}
	return nil
}

// Helper methods for logging

// LogAction logs a general action
func (s *auditService) LogAction(userID *uint, action, resource string, resourceID *uint, resourceName, message string, oldValues, newValues, changes, metadata map[string]interface{}, tags []string, severity string, targetUserID *uint, ipAddress, userAgent, referer, sessionID string) error {
	req := &model.CreateAuditLogRequest{
		UserID:       userID,
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Operation:    s.getOperationFromAction(action),
		Status:       model.StatusSuccess,
		Message:      message,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Referer:      referer,
		SessionID:    sessionID,
		OldValues:    oldValues,
		NewValues:    newValues,
		Changes:      changes,
		Metadata:     metadata,
		Tags:         tags,
		Severity:     severity,
		TargetUserID: targetUserID,
	}

	_, err := s.CreateAuditLog(req)
	return err
}

// LogUserAction logs a user action
func (s *auditService) LogUserAction(userID uint, action, resource string, resourceID *uint, resourceName, message string, oldValues, newValues, changes, metadata map[string]interface{}, tags []string, severity string, targetUserID *uint, ipAddress, userAgent, referer, sessionID string) error {
	return s.LogAction(&userID, action, resource, resourceID, resourceName, message, oldValues, newValues, changes, metadata, tags, severity, targetUserID, ipAddress, userAgent, referer, sessionID)
}

// LogSystemAction logs a system action
func (s *auditService) LogSystemAction(action, resource string, resourceID *uint, resourceName, message string, oldValues, newValues, changes, metadata map[string]interface{}, tags []string, severity string, ipAddress, userAgent, referer, sessionID string) error {
	return s.LogAction(nil, action, resource, resourceID, resourceName, message, oldValues, newValues, changes, metadata, tags, severity, nil, ipAddress, userAgent, referer, sessionID)
}

// Helper methods for response conversion

// toAuditLogResponse converts AuditLog to AuditLogResponse
func (s *auditService) toAuditLogResponse(log *model.AuditLog) *model.AuditLogResponse {
	// Parse JSON fields (simplified - would need proper JSON unmarshaling)
	var oldValues map[string]interface{}
	var newValues map[string]interface{}
	var changes map[string]interface{}
	var metadata map[string]interface{}
	var tags []string

	// Convert User to User (simplified)
	var user *model.User
	if log.User != nil && log.User.ID != 0 {
		user = log.User
	}

	var targetUser *model.User
	if log.TargetUser != nil && log.TargetUser.ID != 0 {
		targetUser = log.TargetUser
	}

	return &model.AuditLogResponse{
		ID:           log.ID,
		UserID:       log.UserID,
		User:         user,
		Action:       log.Action,
		Resource:     log.Resource,
		ResourceID:   log.ResourceID,
		ResourceName: log.ResourceName,
		Operation:    log.Operation,
		Status:       log.Status,
		Message:      log.Message,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		Referer:      log.Referer,
		SessionID:    log.SessionID,
		OldValues:    oldValues,
		NewValues:    newValues,
		Changes:      changes,
		Metadata:     metadata,
		Tags:         tags,
		Severity:     log.Severity,
		TargetUserID: log.TargetUserID,
		TargetUser:   targetUser,
		CreatedAt:    log.CreatedAt,
		UpdatedAt:    log.UpdatedAt,
	}
}

// toAuditConfigResponse converts AuditLogConfig to AuditLogConfigResponse
func (s *auditService) toAuditConfigResponse(config *model.AuditLogConfig) *model.AuditLogConfigResponse {
	// Parse JSON fields (simplified - would need proper JSON unmarshaling)
	var resources []string
	var actions []string
	var excludeUsers []uint

	return &model.AuditLogConfigResponse{
		ID:            config.ID,
		Name:          config.Name,
		Description:   config.Description,
		IsEnabled:     config.IsEnabled,
		LogLevel:      config.LogLevel,
		Resources:     resources,
		Actions:       actions,
		ExcludeUsers:  excludeUsers,
		RetentionDays: config.RetentionDays,
		MaxLogSize:    config.MaxLogSize,
		CreatedAt:     config.CreatedAt,
		UpdatedAt:     config.UpdatedAt,
	}
}

// toAuditSummaryResponse converts AuditLogSummary to AuditLogSummaryResponse
func (s *auditService) toAuditSummaryResponse(summary *model.AuditLogSummary) *model.AuditLogSummaryResponse {
	return &model.AuditLogSummaryResponse{
		ID:           summary.ID,
		Date:         summary.Date,
		Resource:     summary.Resource,
		Action:       summary.Action,
		Status:       summary.Status,
		TotalCount:   summary.TotalCount,
		SuccessCount: summary.SuccessCount,
		FailureCount: summary.FailureCount,
		ErrorCount:   summary.ErrorCount,
		UniqueUsers:  summary.UniqueUsers,
		CreatedAt:    summary.CreatedAt,
		UpdatedAt:    summary.UpdatedAt,
	}
}

// getOperationFromAction determines operation from action
func (s *auditService) getOperationFromAction(action string) string {
	switch action {
	case model.ActionCreate, model.ActionRegister, model.ActionFileUpload, model.ActionOrderCreate, model.ActionProductCreate, model.ActionCategoryCreate, model.ActionBrandCreate, model.ActionReviewCreate, model.ActionCouponCreate, model.ActionDashboardCreate:
		return model.OperationCreate
	case model.ActionRead, model.ActionSearch:
		return model.OperationRead
	case model.ActionUpdate, model.ActionPasswordChange, model.ActionOrderUpdate, model.ActionProductUpdate, model.ActionCategoryUpdate, model.ActionBrandUpdate, model.ActionReviewUpdate, model.ActionCouponUpdate, model.ActionDashboardUpdate, model.ActionInventoryUpdate:
		return model.OperationUpdate
	case model.ActionDelete, model.ActionFileDelete, model.ActionOrderCancel, model.ActionProductDelete, model.ActionCategoryDelete, model.ActionBrandDelete, model.ActionReviewDelete, model.ActionCouponDelete, model.ActionDashboardDelete:
		return model.OperationDelete
	case model.ActionLogin:
		return model.OperationLogin
	case model.ActionLogout:
		return model.OperationLogout
	case model.ActionPermissionGrant, model.ActionRoleAssign:
		return model.OperationGrant
	case model.ActionPermissionRevoke, model.ActionRoleRemove:
		return model.OperationRevoke
	case model.ActionExport:
		return model.OperationExport
	case model.ActionImport:
		return model.OperationImport
	case model.ActionBackup:
		return model.OperationBackup
	case model.ActionRestore:
		return model.OperationRestore
	case model.ActionSystemConfig:
		return model.OperationConfig
	default:
		return model.OperationRead
	}
}
