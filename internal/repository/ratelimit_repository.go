package repository

import (
	"go_app/internal/model"

	"gorm.io/gorm"
)

type RateLimitRepository interface {
	// Rate Limit Rules
	CreateRateLimitRule(rule *model.RateLimitRule) error
	GetRateLimitRuleByID(id uint) (*model.RateLimitRule, error)
	GetRateLimitRuleByName(name string) (*model.RateLimitRule, error)
	GetAllRateLimitRules() ([]model.RateLimitRule, error)
	GetActiveRateLimitRules() ([]model.RateLimitRule, error)
	GetRateLimitRulesByTargetType(targetType string) ([]model.RateLimitRule, error)
	GetRateLimitRulesByScope(scope string) ([]model.RateLimitRule, error)
	UpdateRateLimitRule(rule *model.RateLimitRule) error
	DeleteRateLimitRule(id uint) error

	// Rate Limit Logs
	CreateRateLimitLog(log *model.RateLimitLog) error
	GetRateLimitLogs(page, limit int, filters map[string]interface{}) ([]model.RateLimitLog, int64, error)
	GetRateLimitLogsByRule(ruleID uint, page, limit int) ([]model.RateLimitLog, int64, error)
	GetRateLimitLogsByClient(clientIP string, page, limit int) ([]model.RateLimitLog, int64, error)
	GetRateLimitLogsByUser(userID uint, page, limit int) ([]model.RateLimitLog, int64, error)
	DeleteOldRateLimitLogs(days int) error

	// Rate Limit Stats
	CreateRateLimitStats(stats *model.RateLimitStats) error
	GetRateLimitStats(ruleID uint, period string, page, limit int) ([]model.RateLimitStats, int64, error)
	GetRateLimitStatsByPeriod(period string, page, limit int) ([]model.RateLimitStats, int64, error)
	DeleteOldRateLimitStats(days int) error

	// Whitelist/Blacklist
	CreateWhitelistEntry(entry *model.RateLimitWhitelist) error
	GetWhitelistEntryByID(id uint) (*model.RateLimitWhitelist, error)
	GetAllWhitelistEntries() ([]model.RateLimitWhitelist, error)
	GetActiveWhitelistEntries() ([]model.RateLimitWhitelist, error)
	GetWhitelistEntriesByType(entryType string) ([]model.RateLimitWhitelist, error)
	IsWhitelisted(entryType, value string) (bool, error)
	UpdateWhitelistEntry(entry *model.RateLimitWhitelist) error
	DeleteWhitelistEntry(id uint) error

	CreateBlacklistEntry(entry *model.RateLimitBlacklist) error
	GetBlacklistEntryByID(id uint) (*model.RateLimitBlacklist, error)
	GetAllBlacklistEntries() ([]model.RateLimitBlacklist, error)
	GetActiveBlacklistEntries() ([]model.RateLimitBlacklist, error)
	GetBlacklistEntriesByType(entryType string) ([]model.RateLimitBlacklist, error)
	IsBlacklisted(entryType, value string) (bool, error)
	UpdateBlacklistEntry(entry *model.RateLimitBlacklist) error
	DeleteBlacklistEntry(id uint) error

	// Rate Limit Config
	CreateRateLimitConfig(config *model.RateLimitConfig) error
	GetRateLimitConfigByID(id uint) (*model.RateLimitConfig, error)
	GetRateLimitConfigByName(name string) (*model.RateLimitConfig, error)
	GetActiveRateLimitConfig() (*model.RateLimitConfig, error)
	UpdateRateLimitConfig(config *model.RateLimitConfig) error
	DeleteRateLimitConfig(id uint) error
}

type rateLimitRepository struct {
	db *gorm.DB
}

func NewRateLimitRepository(db *gorm.DB) RateLimitRepository {
	return &rateLimitRepository{db: db}
}

// Rate Limit Rules
func (r *rateLimitRepository) CreateRateLimitRule(rule *model.RateLimitRule) error {
	return r.db.Create(rule).Error
}

func (r *rateLimitRepository) GetRateLimitRuleByID(id uint) (*model.RateLimitRule, error) {
	var rule model.RateLimitRule
	err := r.db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *rateLimitRepository) GetRateLimitRuleByName(name string) (*model.RateLimitRule, error) {
	var rule model.RateLimitRule
	err := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *rateLimitRepository) GetAllRateLimitRules() ([]model.RateLimitRule, error) {
	var rules []model.RateLimitRule
	err := r.db.Where("deleted_at IS NULL").Order("priority DESC, created_at ASC").Find(&rules).Error
	return rules, err
}

func (r *rateLimitRepository) GetActiveRateLimitRules() ([]model.RateLimitRule, error) {
	var rules []model.RateLimitRule
	err := r.db.Where("is_active = ? AND deleted_at IS NULL", true).Order("priority DESC, created_at ASC").Find(&rules).Error
	return rules, err
}

func (r *rateLimitRepository) GetRateLimitRulesByTargetType(targetType string) ([]model.RateLimitRule, error) {
	var rules []model.RateLimitRule
	err := r.db.Where("target_type = ? AND is_active = ? AND deleted_at IS NULL", targetType, true).Order("priority DESC").Find(&rules).Error
	return rules, err
}

func (r *rateLimitRepository) GetRateLimitRulesByScope(scope string) ([]model.RateLimitRule, error) {
	var rules []model.RateLimitRule
	err := r.db.Where("scope = ? AND is_active = ? AND deleted_at IS NULL", scope, true).Order("priority DESC").Find(&rules).Error
	return rules, err
}

func (r *rateLimitRepository) UpdateRateLimitRule(rule *model.RateLimitRule) error {
	return r.db.Save(rule).Error
}

func (r *rateLimitRepository) DeleteRateLimitRule(id uint) error {
	return r.db.Delete(&model.RateLimitRule{}, id).Error
}

// Rate Limit Logs
func (r *rateLimitRepository) CreateRateLimitLog(log *model.RateLimitLog) error {
	return r.db.Create(log).Error
}

func (r *rateLimitRepository) GetRateLimitLogs(page, limit int, filters map[string]interface{}) ([]model.RateLimitLog, int64, error) {
	var logs []model.RateLimitLog
	var total int64

	query := r.db.Model(&model.RateLimitLog{}).Preload("Rule").Preload("User")

	// Apply filters
	if ruleID, ok := filters["rule_id"]; ok {
		query = query.Where("rule_id = ?", ruleID)
	}
	if clientIP, ok := filters["client_ip"]; ok {
		query = query.Where("client_ip = ?", clientIP)
	}
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if violationType, ok := filters["violation_type"]; ok {
		query = query.Where("violation_type = ?", violationType)
	}
	if isBlocked, ok := filters["is_blocked"]; ok {
		query = query.Where("is_blocked = ?", isBlocked)
	}
	if fromDate, ok := filters["from_date"]; ok {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate, ok := filters["to_date"]; ok {
		query = query.Where("created_at <= ?", toDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *rateLimitRepository) GetRateLimitLogsByRule(ruleID uint, page, limit int) ([]model.RateLimitLog, int64, error) {
	var logs []model.RateLimitLog
	var total int64

	query := r.db.Model(&model.RateLimitLog{}).Where("rule_id = ?", ruleID).Preload("Rule").Preload("User")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *rateLimitRepository) GetRateLimitLogsByClient(clientIP string, page, limit int) ([]model.RateLimitLog, int64, error) {
	var logs []model.RateLimitLog
	var total int64

	query := r.db.Model(&model.RateLimitLog{}).Where("client_ip = ?", clientIP).Preload("Rule").Preload("User")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *rateLimitRepository) GetRateLimitLogsByUser(userID uint, page, limit int) ([]model.RateLimitLog, int64, error) {
	var logs []model.RateLimitLog
	var total int64

	query := r.db.Model(&model.RateLimitLog{}).Where("user_id = ?", userID).Preload("Rule").Preload("User")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *rateLimitRepository) DeleteOldRateLimitLogs(days int) error {
	return r.db.Where("created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", days).Delete(&model.RateLimitLog{}).Error
}

// Rate Limit Stats
func (r *rateLimitRepository) CreateRateLimitStats(stats *model.RateLimitStats) error {
	return r.db.Create(stats).Error
}

func (r *rateLimitRepository) GetRateLimitStats(ruleID uint, period string, page, limit int) ([]model.RateLimitStats, int64, error) {
	var stats []model.RateLimitStats
	var total int64

	query := r.db.Model(&model.RateLimitStats{}).Where("rule_id = ?", ruleID).Preload("Rule")

	if period != "" {
		query = query.Where("period = ?", period)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("period_start DESC").Find(&stats).Error
	return stats, total, err
}

func (r *rateLimitRepository) GetRateLimitStatsByPeriod(period string, page, limit int) ([]model.RateLimitStats, int64, error) {
	var stats []model.RateLimitStats
	var total int64

	query := r.db.Model(&model.RateLimitStats{}).Preload("Rule")

	if period != "" {
		query = query.Where("period = ?", period)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("period_start DESC").Find(&stats).Error
	return stats, total, err
}

func (r *rateLimitRepository) DeleteOldRateLimitStats(days int) error {
	return r.db.Where("created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", days).Delete(&model.RateLimitStats{}).Error
}

// Whitelist/Blacklist
func (r *rateLimitRepository) CreateWhitelistEntry(entry *model.RateLimitWhitelist) error {
	return r.db.Create(entry).Error
}

func (r *rateLimitRepository) GetWhitelistEntryByID(id uint) (*model.RateLimitWhitelist, error) {
	var entry model.RateLimitWhitelist
	err := r.db.Where("id = ?", id).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *rateLimitRepository) GetAllWhitelistEntries() ([]model.RateLimitWhitelist, error) {
	var entries []model.RateLimitWhitelist
	err := r.db.Where("deleted_at IS NULL").Order("created_at DESC").Find(&entries).Error
	return entries, err
}

func (r *rateLimitRepository) GetActiveWhitelistEntries() ([]model.RateLimitWhitelist, error) {
	var entries []model.RateLimitWhitelist
	err := r.db.Where("is_active = ? AND deleted_at IS NULL", true).Order("created_at DESC").Find(&entries).Error
	return entries, err
}

func (r *rateLimitRepository) GetWhitelistEntriesByType(entryType string) ([]model.RateLimitWhitelist, error) {
	var entries []model.RateLimitWhitelist
	err := r.db.Where("type = ? AND is_active = ? AND deleted_at IS NULL", entryType, true).Find(&entries).Error
	return entries, err
}

func (r *rateLimitRepository) IsWhitelisted(entryType, value string) (bool, error) {
	var count int64
	err := r.db.Model(&model.RateLimitWhitelist{}).
		Where("type = ? AND value = ? AND is_active = ? AND deleted_at IS NULL", entryType, value, true).
		Count(&count).Error
	return count > 0, err
}

func (r *rateLimitRepository) UpdateWhitelistEntry(entry *model.RateLimitWhitelist) error {
	return r.db.Save(entry).Error
}

func (r *rateLimitRepository) DeleteWhitelistEntry(id uint) error {
	return r.db.Delete(&model.RateLimitWhitelist{}, id).Error
}

func (r *rateLimitRepository) CreateBlacklistEntry(entry *model.RateLimitBlacklist) error {
	return r.db.Create(entry).Error
}

func (r *rateLimitRepository) GetBlacklistEntryByID(id uint) (*model.RateLimitBlacklist, error) {
	var entry model.RateLimitBlacklist
	err := r.db.Where("id = ?", id).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *rateLimitRepository) GetAllBlacklistEntries() ([]model.RateLimitBlacklist, error) {
	var entries []model.RateLimitBlacklist
	err := r.db.Where("deleted_at IS NULL").Order("created_at DESC").Find(&entries).Error
	return entries, err
}

func (r *rateLimitRepository) GetActiveBlacklistEntries() ([]model.RateLimitBlacklist, error) {
	var entries []model.RateLimitBlacklist
	err := r.db.Where("is_active = ? AND deleted_at IS NULL", true).Order("created_at DESC").Find(&entries).Error
	return entries, err
}

func (r *rateLimitRepository) GetBlacklistEntriesByType(entryType string) ([]model.RateLimitBlacklist, error) {
	var entries []model.RateLimitBlacklist
	err := r.db.Where("type = ? AND is_active = ? AND deleted_at IS NULL", entryType, true).Find(&entries).Error
	return entries, err
}

func (r *rateLimitRepository) IsBlacklisted(entryType, value string) (bool, error) {
	var count int64
	err := r.db.Model(&model.RateLimitBlacklist{}).
		Where("type = ? AND value = ? AND is_active = ? AND deleted_at IS NULL", entryType, value, true).
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Count(&count).Error
	return count > 0, err
}

func (r *rateLimitRepository) UpdateBlacklistEntry(entry *model.RateLimitBlacklist) error {
	return r.db.Save(entry).Error
}

func (r *rateLimitRepository) DeleteBlacklistEntry(id uint) error {
	return r.db.Delete(&model.RateLimitBlacklist{}, id).Error
}

// Rate Limit Config
func (r *rateLimitRepository) CreateRateLimitConfig(config *model.RateLimitConfig) error {
	return r.db.Create(config).Error
}

func (r *rateLimitRepository) GetRateLimitConfigByID(id uint) (*model.RateLimitConfig, error) {
	var config model.RateLimitConfig
	err := r.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *rateLimitRepository) GetRateLimitConfigByName(name string) (*model.RateLimitConfig, error) {
	var config model.RateLimitConfig
	err := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *rateLimitRepository) GetActiveRateLimitConfig() (*model.RateLimitConfig, error) {
	var config model.RateLimitConfig
	err := r.db.Where("is_enabled = ? AND deleted_at IS NULL", true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *rateLimitRepository) UpdateRateLimitConfig(config *model.RateLimitConfig) error {
	return r.db.Save(config).Error
}

func (r *rateLimitRepository) DeleteRateLimitConfig(id uint) error {
	return r.db.Delete(&model.RateLimitConfig{}, id).Error
}
