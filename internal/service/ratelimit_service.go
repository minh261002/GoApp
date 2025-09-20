package service

import (
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"go_app/pkg/ratelimit"
)

type RateLimitService interface {
	// Rate Limit Rules
	CreateRateLimitRule(req *model.RateLimitRuleRequest) (*model.RateLimitRuleResponse, error)
	GetRateLimitRuleByID(id uint) (*model.RateLimitRuleResponse, error)
	GetRateLimitRuleByName(name string) (*model.RateLimitRuleResponse, error)
	GetAllRateLimitRules() ([]model.RateLimitRuleResponse, error)
	GetActiveRateLimitRules() ([]model.RateLimitRuleResponse, error)
	UpdateRateLimitRule(id uint, req *model.RateLimitRuleRequest) (*model.RateLimitRuleResponse, error)
	DeleteRateLimitRule(id uint) error

	// Rate Limit Logs
	GetRateLimitLogs(page, limit int, filters map[string]interface{}) ([]model.RateLimitLogResponse, int64, error)
	GetRateLimitLogsByRule(ruleID uint, page, limit int) ([]model.RateLimitLogResponse, int64, error)
	GetRateLimitLogsByClient(clientIP string, page, limit int) ([]model.RateLimitLogResponse, int64, error)
	GetRateLimitLogsByUser(userID uint, page, limit int) ([]model.RateLimitLogResponse, int64, error)
	CleanupOldLogs(days int) error

	// Rate Limit Stats
	GetRateLimitStats(ruleID uint, period string, page, limit int) ([]model.RateLimitStatsResponse, int64, error)
	GetRateLimitStatsByPeriod(period string, page, limit int) ([]model.RateLimitStatsResponse, int64, error)
	CleanupOldStats(days int) error

	// Whitelist/Blacklist
	CreateWhitelistEntry(req *model.RateLimitWhitelistRequest) (*model.RateLimitWhitelist, error)
	GetWhitelistEntryByID(id uint) (*model.RateLimitWhitelist, error)
	GetAllWhitelistEntries() ([]model.RateLimitWhitelist, error)
	GetActiveWhitelistEntries() ([]model.RateLimitWhitelist, error)
	IsWhitelisted(entryType, value string) (bool, error)
	UpdateWhitelistEntry(id uint, req *model.RateLimitWhitelistRequest) (*model.RateLimitWhitelist, error)
	DeleteWhitelistEntry(id uint) error

	CreateBlacklistEntry(req *model.RateLimitBlacklistRequest) (*model.RateLimitBlacklist, error)
	GetBlacklistEntryByID(id uint) (*model.RateLimitBlacklist, error)
	GetAllBlacklistEntries() ([]model.RateLimitBlacklist, error)
	GetActiveBlacklistEntries() ([]model.RateLimitBlacklist, error)
	IsBlacklisted(entryType, value string) (bool, error)
	UpdateBlacklistEntry(id uint, req *model.RateLimitBlacklistRequest) (*model.RateLimitBlacklist, error)
	DeleteBlacklistEntry(id uint) error

	// Rate Limit Config
	CreateRateLimitConfig(req *model.RateLimitConfigRequest) (*model.RateLimitConfig, error)
	GetRateLimitConfigByID(id uint) (*model.RateLimitConfig, error)
	GetActiveRateLimitConfig() (*model.RateLimitConfig, error)
	UpdateRateLimitConfig(id uint, req *model.RateLimitConfigRequest) (*model.RateLimitConfig, error)
	DeleteRateLimitConfig(id uint) error

	// Rate Limit Manager
	GetRateLimitManager() *ratelimit.RateLimitManager
	GetRateLimitInfo(ruleName, clientID string) (*ratelimit.RateLimitInfo, error)
	ClearRateLimit(ruleName, clientID string) error
}

type rateLimitService struct {
	rateLimitRepo repository.RateLimitRepository
	manager       *ratelimit.RateLimitManager
}

func NewRateLimitService(rateLimitRepo repository.RateLimitRepository, redisClient interface{}) RateLimitService {
	// Convert redisClient to *redis.Client if needed
	// This is a placeholder - you'll need to pass the actual Redis client
	manager := ratelimit.NewRateLimitManager(nil) // Will be set properly in router

	return &rateLimitService{
		rateLimitRepo: rateLimitRepo,
		manager:       manager,
	}
}

// Rate Limit Rules
func (s *rateLimitService) CreateRateLimitRule(req *model.RateLimitRuleRequest) (*model.RateLimitRuleResponse, error) {
	rule := &model.RateLimitRule{
		Name:         req.Name,
		Description:  req.Description,
		Requests:     req.Requests,
		Window:       req.Window,
		WindowType:   req.WindowType,
		TargetType:   req.TargetType,
		TargetValue:  req.TargetValue,
		Scope:        req.Scope,
		ScopeValue:   req.ScopeValue,
		IsActive:     req.IsActive,
		Priority:     req.Priority,
		ErrorCode:    req.ErrorCode,
		ErrorMessage: req.ErrorMessage,
	}

	if rule.ErrorCode == 0 {
		rule.ErrorCode = 429
	}

	if err := s.rateLimitRepo.CreateRateLimitRule(rule); err != nil {
		logger.Errorf("Failed to create rate limit rule: %v", err)
		return nil, fmt.Errorf("failed to create rate limit rule")
	}

	return s.toRateLimitRuleResponse(rule), nil
}

func (s *rateLimitService) GetRateLimitRuleByID(id uint) (*model.RateLimitRuleResponse, error) {
	rule, err := s.rateLimitRepo.GetRateLimitRuleByID(id)
	if err != nil {
		logger.Errorf("Failed to get rate limit rule by ID %d: %v", id, err)
		return nil, fmt.Errorf("rate limit rule not found")
	}
	return s.toRateLimitRuleResponse(rule), nil
}

func (s *rateLimitService) GetRateLimitRuleByName(name string) (*model.RateLimitRuleResponse, error) {
	rule, err := s.rateLimitRepo.GetRateLimitRuleByName(name)
	if err != nil {
		logger.Errorf("Failed to get rate limit rule by name %s: %v", name, err)
		return nil, fmt.Errorf("rate limit rule not found")
	}
	return s.toRateLimitRuleResponse(rule), nil
}

func (s *rateLimitService) GetAllRateLimitRules() ([]model.RateLimitRuleResponse, error) {
	rules, err := s.rateLimitRepo.GetAllRateLimitRules()
	if err != nil {
		logger.Errorf("Failed to get all rate limit rules: %v", err)
		return nil, fmt.Errorf("failed to get rate limit rules")
	}

	var responses []model.RateLimitRuleResponse
	for _, rule := range rules {
		responses = append(responses, *s.toRateLimitRuleResponse(&rule))
	}
	return responses, nil
}

func (s *rateLimitService) GetActiveRateLimitRules() ([]model.RateLimitRuleResponse, error) {
	rules, err := s.rateLimitRepo.GetActiveRateLimitRules()
	if err != nil {
		logger.Errorf("Failed to get active rate limit rules: %v", err)
		return nil, fmt.Errorf("failed to get active rate limit rules")
	}

	var responses []model.RateLimitRuleResponse
	for _, rule := range rules {
		responses = append(responses, *s.toRateLimitRuleResponse(&rule))
	}
	return responses, nil
}

func (s *rateLimitService) UpdateRateLimitRule(id uint, req *model.RateLimitRuleRequest) (*model.RateLimitRuleResponse, error) {
	rule, err := s.rateLimitRepo.GetRateLimitRuleByID(id)
	if err != nil {
		logger.Errorf("Failed to get rate limit rule %d: %v", id, err)
		return nil, fmt.Errorf("rate limit rule not found")
	}

	rule.Name = req.Name
	rule.Description = req.Description
	rule.Requests = req.Requests
	rule.Window = req.Window
	rule.WindowType = req.WindowType
	rule.TargetType = req.TargetType
	rule.TargetValue = req.TargetValue
	rule.Scope = req.Scope
	rule.ScopeValue = req.ScopeValue
	rule.IsActive = req.IsActive
	rule.Priority = req.Priority
	rule.ErrorCode = req.ErrorCode
	rule.ErrorMessage = req.ErrorMessage

	if err := s.rateLimitRepo.UpdateRateLimitRule(rule); err != nil {
		logger.Errorf("Failed to update rate limit rule %d: %v", id, err)
		return nil, fmt.Errorf("failed to update rate limit rule")
	}

	return s.toRateLimitRuleResponse(rule), nil
}

func (s *rateLimitService) DeleteRateLimitRule(id uint) error {
	if err := s.rateLimitRepo.DeleteRateLimitRule(id); err != nil {
		logger.Errorf("Failed to delete rate limit rule %d: %v", id, err)
		return fmt.Errorf("failed to delete rate limit rule")
	}
	return nil
}

// Rate Limit Logs
func (s *rateLimitService) GetRateLimitLogs(page, limit int, filters map[string]interface{}) ([]model.RateLimitLogResponse, int64, error) {
	logs, total, err := s.rateLimitRepo.GetRateLimitLogs(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get rate limit logs: %v", err)
		return nil, 0, fmt.Errorf("failed to get rate limit logs")
	}

	var responses []model.RateLimitLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toRateLimitLogResponse(&log))
	}
	return responses, total, nil
}

func (s *rateLimitService) GetRateLimitLogsByRule(ruleID uint, page, limit int) ([]model.RateLimitLogResponse, int64, error) {
	logs, total, err := s.rateLimitRepo.GetRateLimitLogsByRule(ruleID, page, limit)
	if err != nil {
		logger.Errorf("Failed to get rate limit logs by rule %d: %v", ruleID, err)
		return nil, 0, fmt.Errorf("failed to get rate limit logs")
	}

	var responses []model.RateLimitLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toRateLimitLogResponse(&log))
	}
	return responses, total, nil
}

func (s *rateLimitService) GetRateLimitLogsByClient(clientIP string, page, limit int) ([]model.RateLimitLogResponse, int64, error) {
	logs, total, err := s.rateLimitRepo.GetRateLimitLogsByClient(clientIP, page, limit)
	if err != nil {
		logger.Errorf("Failed to get rate limit logs by client %s: %v", clientIP, err)
		return nil, 0, fmt.Errorf("failed to get rate limit logs")
	}

	var responses []model.RateLimitLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toRateLimitLogResponse(&log))
	}
	return responses, total, nil
}

func (s *rateLimitService) GetRateLimitLogsByUser(userID uint, page, limit int) ([]model.RateLimitLogResponse, int64, error) {
	logs, total, err := s.rateLimitRepo.GetRateLimitLogsByUser(userID, page, limit)
	if err != nil {
		logger.Errorf("Failed to get rate limit logs by user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to get rate limit logs")
	}

	var responses []model.RateLimitLogResponse
	for _, log := range logs {
		responses = append(responses, *s.toRateLimitLogResponse(&log))
	}
	return responses, total, nil
}

func (s *rateLimitService) CleanupOldLogs(days int) error {
	if err := s.rateLimitRepo.DeleteOldRateLimitLogs(days); err != nil {
		logger.Errorf("Failed to cleanup old rate limit logs: %v", err)
		return fmt.Errorf("failed to cleanup old logs")
	}
	return nil
}

// Rate Limit Stats
func (s *rateLimitService) GetRateLimitStats(ruleID uint, period string, page, limit int) ([]model.RateLimitStatsResponse, int64, error) {
	stats, total, err := s.rateLimitRepo.GetRateLimitStats(ruleID, period, page, limit)
	if err != nil {
		logger.Errorf("Failed to get rate limit stats: %v", err)
		return nil, 0, fmt.Errorf("failed to get rate limit stats")
	}

	var responses []model.RateLimitStatsResponse
	for _, stat := range stats {
		responses = append(responses, *s.toRateLimitStatsResponse(&stat))
	}
	return responses, total, nil
}

func (s *rateLimitService) GetRateLimitStatsByPeriod(period string, page, limit int) ([]model.RateLimitStatsResponse, int64, error) {
	stats, total, err := s.rateLimitRepo.GetRateLimitStatsByPeriod(period, page, limit)
	if err != nil {
		logger.Errorf("Failed to get rate limit stats by period: %v", err)
		return nil, 0, fmt.Errorf("failed to get rate limit stats")
	}

	var responses []model.RateLimitStatsResponse
	for _, stat := range stats {
		responses = append(responses, *s.toRateLimitStatsResponse(&stat))
	}
	return responses, total, nil
}

func (s *rateLimitService) CleanupOldStats(days int) error {
	if err := s.rateLimitRepo.DeleteOldRateLimitStats(days); err != nil {
		logger.Errorf("Failed to cleanup old rate limit stats: %v", err)
		return fmt.Errorf("failed to cleanup old stats")
	}
	return nil
}

// Whitelist/Blacklist
func (s *rateLimitService) CreateWhitelistEntry(req *model.RateLimitWhitelistRequest) (*model.RateLimitWhitelist, error) {
	entry := &model.RateLimitWhitelist{
		Type:        req.Type,
		Value:       req.Value,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := s.rateLimitRepo.CreateWhitelistEntry(entry); err != nil {
		logger.Errorf("Failed to create whitelist entry: %v", err)
		return nil, fmt.Errorf("failed to create whitelist entry")
	}

	return entry, nil
}

func (s *rateLimitService) GetWhitelistEntryByID(id uint) (*model.RateLimitWhitelist, error) {
	entry, err := s.rateLimitRepo.GetWhitelistEntryByID(id)
	if err != nil {
		logger.Errorf("Failed to get whitelist entry %d: %v", id, err)
		return nil, fmt.Errorf("whitelist entry not found")
	}
	return entry, nil
}

func (s *rateLimitService) GetAllWhitelistEntries() ([]model.RateLimitWhitelist, error) {
	entries, err := s.rateLimitRepo.GetAllWhitelistEntries()
	if err != nil {
		logger.Errorf("Failed to get all whitelist entries: %v", err)
		return nil, fmt.Errorf("failed to get whitelist entries")
	}
	return entries, nil
}

func (s *rateLimitService) GetActiveWhitelistEntries() ([]model.RateLimitWhitelist, error) {
	entries, err := s.rateLimitRepo.GetActiveWhitelistEntries()
	if err != nil {
		logger.Errorf("Failed to get active whitelist entries: %v", err)
		return nil, fmt.Errorf("failed to get active whitelist entries")
	}
	return entries, nil
}

func (s *rateLimitService) IsWhitelisted(entryType, value string) (bool, error) {
	return s.rateLimitRepo.IsWhitelisted(entryType, value)
}

func (s *rateLimitService) UpdateWhitelistEntry(id uint, req *model.RateLimitWhitelistRequest) (*model.RateLimitWhitelist, error) {
	entry, err := s.rateLimitRepo.GetWhitelistEntryByID(id)
	if err != nil {
		logger.Errorf("Failed to get whitelist entry %d: %v", id, err)
		return nil, fmt.Errorf("whitelist entry not found")
	}

	entry.Type = req.Type
	entry.Value = req.Value
	entry.Description = req.Description
	entry.IsActive = req.IsActive

	if err := s.rateLimitRepo.UpdateWhitelistEntry(entry); err != nil {
		logger.Errorf("Failed to update whitelist entry %d: %v", id, err)
		return nil, fmt.Errorf("failed to update whitelist entry")
	}

	return entry, nil
}

func (s *rateLimitService) DeleteWhitelistEntry(id uint) error {
	if err := s.rateLimitRepo.DeleteWhitelistEntry(id); err != nil {
		logger.Errorf("Failed to delete whitelist entry %d: %v", id, err)
		return fmt.Errorf("failed to delete whitelist entry")
	}
	return nil
}

func (s *rateLimitService) CreateBlacklistEntry(req *model.RateLimitBlacklistRequest) (*model.RateLimitBlacklist, error) {
	entry := &model.RateLimitBlacklist{
		Type:      req.Type,
		Value:     req.Value,
		Reason:    req.Reason,
		IsActive:  req.IsActive,
		ExpiresAt: req.ExpiresAt,
	}

	if err := s.rateLimitRepo.CreateBlacklistEntry(entry); err != nil {
		logger.Errorf("Failed to create blacklist entry: %v", err)
		return nil, fmt.Errorf("failed to create blacklist entry")
	}

	return entry, nil
}

func (s *rateLimitService) GetBlacklistEntryByID(id uint) (*model.RateLimitBlacklist, error) {
	entry, err := s.rateLimitRepo.GetBlacklistEntryByID(id)
	if err != nil {
		logger.Errorf("Failed to get blacklist entry %d: %v", id, err)
		return nil, fmt.Errorf("blacklist entry not found")
	}
	return entry, nil
}

func (s *rateLimitService) GetAllBlacklistEntries() ([]model.RateLimitBlacklist, error) {
	entries, err := s.rateLimitRepo.GetAllBlacklistEntries()
	if err != nil {
		logger.Errorf("Failed to get all blacklist entries: %v", err)
		return nil, fmt.Errorf("failed to get blacklist entries")
	}
	return entries, nil
}

func (s *rateLimitService) GetActiveBlacklistEntries() ([]model.RateLimitBlacklist, error) {
	entries, err := s.rateLimitRepo.GetActiveBlacklistEntries()
	if err != nil {
		logger.Errorf("Failed to get active blacklist entries: %v", err)
		return nil, fmt.Errorf("failed to get active blacklist entries")
	}
	return entries, nil
}

func (s *rateLimitService) IsBlacklisted(entryType, value string) (bool, error) {
	return s.rateLimitRepo.IsBlacklisted(entryType, value)
}

func (s *rateLimitService) UpdateBlacklistEntry(id uint, req *model.RateLimitBlacklistRequest) (*model.RateLimitBlacklist, error) {
	entry, err := s.rateLimitRepo.GetBlacklistEntryByID(id)
	if err != nil {
		logger.Errorf("Failed to get blacklist entry %d: %v", id, err)
		return nil, fmt.Errorf("blacklist entry not found")
	}

	entry.Type = req.Type
	entry.Value = req.Value
	entry.Reason = req.Reason
	entry.IsActive = req.IsActive
	entry.ExpiresAt = req.ExpiresAt

	if err := s.rateLimitRepo.UpdateBlacklistEntry(entry); err != nil {
		logger.Errorf("Failed to update blacklist entry %d: %v", id, err)
		return nil, fmt.Errorf("failed to update blacklist entry")
	}

	return entry, nil
}

func (s *rateLimitService) DeleteBlacklistEntry(id uint) error {
	if err := s.rateLimitRepo.DeleteBlacklistEntry(id); err != nil {
		logger.Errorf("Failed to delete blacklist entry %d: %v", id, err)
		return fmt.Errorf("failed to delete blacklist entry")
	}
	return nil
}

// Rate Limit Config
func (s *rateLimitService) CreateRateLimitConfig(req *model.RateLimitConfigRequest) (*model.RateLimitConfig, error) {
	config := &model.RateLimitConfig{
		Name:               req.Name,
		Description:        req.Description,
		IsEnabled:          req.IsEnabled,
		DefaultRule:        req.DefaultRule,
		RedisHost:          req.RedisHost,
		RedisPort:          req.RedisPort,
		RedisDB:            req.RedisDB,
		RedisPassword:      req.RedisPassword,
		LogRetentionDays:   req.LogRetentionDays,
		StatsRetentionDays: req.StatsRetentionDays,
	}

	if err := s.rateLimitRepo.CreateRateLimitConfig(config); err != nil {
		logger.Errorf("Failed to create rate limit config: %v", err)
		return nil, fmt.Errorf("failed to create rate limit config")
	}

	return config, nil
}

func (s *rateLimitService) GetRateLimitConfigByID(id uint) (*model.RateLimitConfig, error) {
	config, err := s.rateLimitRepo.GetRateLimitConfigByID(id)
	if err != nil {
		logger.Errorf("Failed to get rate limit config %d: %v", id, err)
		return nil, fmt.Errorf("rate limit config not found")
	}
	return config, nil
}

func (s *rateLimitService) GetActiveRateLimitConfig() (*model.RateLimitConfig, error) {
	config, err := s.rateLimitRepo.GetActiveRateLimitConfig()
	if err != nil {
		logger.Errorf("Failed to get active rate limit config: %v", err)
		return nil, fmt.Errorf("active rate limit config not found")
	}
	return config, nil
}

func (s *rateLimitService) UpdateRateLimitConfig(id uint, req *model.RateLimitConfigRequest) (*model.RateLimitConfig, error) {
	config, err := s.rateLimitRepo.GetRateLimitConfigByID(id)
	if err != nil {
		logger.Errorf("Failed to get rate limit config %d: %v", id, err)
		return nil, fmt.Errorf("rate limit config not found")
	}

	config.Name = req.Name
	config.Description = req.Description
	config.IsEnabled = req.IsEnabled
	config.DefaultRule = req.DefaultRule
	config.RedisHost = req.RedisHost
	config.RedisPort = req.RedisPort
	config.RedisDB = req.RedisDB
	config.RedisPassword = req.RedisPassword
	config.LogRetentionDays = req.LogRetentionDays
	config.StatsRetentionDays = req.StatsRetentionDays

	if err := s.rateLimitRepo.UpdateRateLimitConfig(config); err != nil {
		logger.Errorf("Failed to update rate limit config %d: %v", id, err)
		return nil, fmt.Errorf("failed to update rate limit config")
	}

	return config, nil
}

func (s *rateLimitService) DeleteRateLimitConfig(id uint) error {
	if err := s.rateLimitRepo.DeleteRateLimitConfig(id); err != nil {
		logger.Errorf("Failed to delete rate limit config %d: %v", id, err)
		return fmt.Errorf("failed to delete rate limit config")
	}
	return nil
}

// Rate Limit Manager
func (s *rateLimitService) GetRateLimitManager() *ratelimit.RateLimitManager {
	return s.manager
}

func (s *rateLimitService) GetRateLimitInfo(ruleName, clientID string) (*ratelimit.RateLimitInfo, error) {
	return s.manager.GetRateLimitInfo(ruleName, clientID)
}

func (s *rateLimitService) ClearRateLimit(ruleName, clientID string) error {
	return s.manager.ClearRateLimit(ruleName, clientID)
}

// Helper methods
func (s *rateLimitService) toRateLimitRuleResponse(rule *model.RateLimitRule) *model.RateLimitRuleResponse {
	return &model.RateLimitRuleResponse{
		ID:           rule.ID,
		Name:         rule.Name,
		Description:  rule.Description,
		Requests:     rule.Requests,
		Window:       rule.Window,
		WindowType:   rule.WindowType,
		TargetType:   rule.TargetType,
		TargetValue:  rule.TargetValue,
		Scope:        rule.Scope,
		ScopeValue:   rule.ScopeValue,
		IsActive:     rule.IsActive,
		Priority:     rule.Priority,
		ErrorCode:    rule.ErrorCode,
		ErrorMessage: rule.ErrorMessage,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	}
}

func (s *rateLimitService) toRateLimitLogResponse(log *model.RateLimitLog) *model.RateLimitLogResponse {
	ruleName := ""
	if log.Rule != nil {
		ruleName = log.Rule.Name
	}

	return &model.RateLimitLogResponse{
		ID:            log.ID,
		RuleID:        log.RuleID,
		RuleName:      ruleName,
		ClientIP:      log.ClientIP,
		UserID:        log.UserID,
		APIKey:        log.APIKey,
		Method:        log.Method,
		Path:          log.Path,
		UserAgent:     log.UserAgent,
		Referer:       log.Referer,
		Limit:         log.Limit,
		Current:       log.Current,
		Remaining:     log.Remaining,
		ResetTime:     log.ResetTime,
		ViolationType: log.ViolationType,
		IsBlocked:     log.IsBlocked,
		CreatedAt:     log.CreatedAt,
	}
}

func (s *rateLimitService) toRateLimitStatsResponse(stat *model.RateLimitStats) *model.RateLimitStatsResponse {
	ruleName := ""
	if stat.Rule != nil {
		ruleName = stat.Rule.Name
	}

	return &model.RateLimitStatsResponse{
		ID:              stat.ID,
		RuleID:          stat.RuleID,
		RuleName:        ruleName,
		Period:          stat.Period,
		PeriodStart:     stat.PeriodStart,
		PeriodEnd:       stat.PeriodEnd,
		TotalRequests:   stat.TotalRequests,
		BlockedRequests: stat.BlockedRequests,
		UniqueClients:   stat.UniqueClients,
		AverageRequests: stat.AverageRequests,
		PeakRequests:    stat.PeakRequests,
		CreatedAt:       stat.CreatedAt,
	}
}
