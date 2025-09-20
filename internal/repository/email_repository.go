package repository

import (
	"go_app/internal/model"

	"gorm.io/gorm"
)

type EmailRepository interface {
	// Email Templates
	CreateEmailTemplate(template *model.EmailTemplate) error
	GetEmailTemplateByID(id uint) (*model.EmailTemplate, error)
	GetEmailTemplateByName(name string) (*model.EmailTemplate, error)
	GetEmailTemplatesByType(templateType string) ([]model.EmailTemplate, error)
	UpdateEmailTemplate(template *model.EmailTemplate) error
	DeleteEmailTemplate(id uint) error
	GetAllEmailTemplates(page, limit int) ([]model.EmailTemplate, int64, error)

	// Email Queue
	CreateEmailQueue(email *model.EmailQueue) error
	GetEmailQueueByID(id uint) (*model.EmailQueue, error)
	GetPendingEmails(limit int) ([]model.EmailQueue, error)
	GetScheduledEmails(limit int) ([]model.EmailQueue, error)
	GetFailedEmails(limit int) ([]model.EmailQueue, error)
	UpdateEmailQueue(email *model.EmailQueue) error
	DeleteEmailQueue(id uint) error
	GetEmailQueueStats() (map[string]interface{}, error)

	// Email Logs
	CreateEmailLog(log *model.EmailLog) error
	GetEmailLogsByQueueID(queueID uint) ([]model.EmailLog, error)
	GetEmailLogs(page, limit int, filters map[string]interface{}) ([]model.EmailLog, int64, error)
	GetEmailStats() (*model.EmailStats, error)

	// Email Config
	CreateEmailConfig(config *model.EmailConfig) error
	GetEmailConfigByID(id uint) (*model.EmailConfig, error)
	GetActiveEmailConfig() (*model.EmailConfig, error)
	UpdateEmailConfig(config *model.EmailConfig) error
	DeleteEmailConfig(id uint) error
	GetAllEmailConfigs() ([]model.EmailConfig, error)
}

type emailRepository struct {
	db *gorm.DB
}

func NewEmailRepository(db *gorm.DB) EmailRepository {
	return &emailRepository{db: db}
}

// Email Templates
func (r *emailRepository) CreateEmailTemplate(template *model.EmailTemplate) error {
	return r.db.Create(template).Error
}

func (r *emailRepository) GetEmailTemplateByID(id uint) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	err := r.db.Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *emailRepository) GetEmailTemplateByName(name string) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	err := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *emailRepository) GetEmailTemplatesByType(templateType string) ([]model.EmailTemplate, error) {
	var templates []model.EmailTemplate
	err := r.db.Where("type = ? AND is_active = ? AND deleted_at IS NULL", templateType, true).Find(&templates).Error
	return templates, err
}

func (r *emailRepository) UpdateEmailTemplate(template *model.EmailTemplate) error {
	return r.db.Save(template).Error
}

func (r *emailRepository) DeleteEmailTemplate(id uint) error {
	return r.db.Delete(&model.EmailTemplate{}, id).Error
}

func (r *emailRepository) GetAllEmailTemplates(page, limit int) ([]model.EmailTemplate, int64, error) {
	var templates []model.EmailTemplate
	var total int64

	query := r.db.Model(&model.EmailTemplate{}).Where("deleted_at IS NULL")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&templates).Error
	return templates, total, err
}

// Email Queue
func (r *emailRepository) CreateEmailQueue(email *model.EmailQueue) error {
	return r.db.Create(email).Error
}

func (r *emailRepository) GetEmailQueueByID(id uint) (*model.EmailQueue, error) {
	var email model.EmailQueue
	err := r.db.Preload("Template").Where("id = ?", id).First(&email).Error
	if err != nil {
		return nil, err
	}
	return &email, nil
}

func (r *emailRepository) GetPendingEmails(limit int) ([]model.EmailQueue, error) {
	var emails []model.EmailQueue
	err := r.db.Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= NOW())", model.EmailStatusPending).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&emails).Error
	return emails, err
}

func (r *emailRepository) GetScheduledEmails(limit int) ([]model.EmailQueue, error) {
	var emails []model.EmailQueue
	err := r.db.Where("status = ? AND scheduled_at > NOW()", model.EmailStatusPending).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&emails).Error
	return emails, err
}

func (r *emailRepository) GetFailedEmails(limit int) ([]model.EmailQueue, error) {
	var emails []model.EmailQueue
	err := r.db.Where("status = ? AND attempts < max_attempts", model.EmailStatusFailed).
		Order("created_at ASC").
		Limit(limit).
		Find(&emails).Error
	return emails, err
}

func (r *emailRepository) UpdateEmailQueue(email *model.EmailQueue) error {
	return r.db.Save(email).Error
}

func (r *emailRepository) DeleteEmailQueue(id uint) error {
	return r.db.Delete(&model.EmailQueue{}, id).Error
}

func (r *emailRepository) GetEmailQueueStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count by status
	var pending, sending, sent, failed int64
	r.db.Model(&model.EmailQueue{}).Where("status = ?", model.EmailStatusPending).Count(&pending)
	r.db.Model(&model.EmailQueue{}).Where("status = ?", model.EmailStatusSending).Count(&sending)
	r.db.Model(&model.EmailQueue{}).Where("status = ?", model.EmailStatusSent).Count(&sent)
	r.db.Model(&model.EmailQueue{}).Where("status = ?", model.EmailStatusFailed).Count(&failed)

	stats["pending"] = pending
	stats["sending"] = sending
	stats["sent"] = sent
	stats["failed"] = failed
	stats["total"] = pending + sending + sent + failed

	// Success rate
	total := pending + sending + sent + failed
	if total > 0 {
		stats["success_rate"] = float64(sent) / float64(total) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	return stats, nil
}

// Email Logs
func (r *emailRepository) CreateEmailLog(log *model.EmailLog) error {
	return r.db.Create(log).Error
}

func (r *emailRepository) GetEmailLogsByQueueID(queueID uint) ([]model.EmailLog, error) {
	var logs []model.EmailLog
	err := r.db.Where("email_queue_id = ?", queueID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *emailRepository) GetEmailLogs(page, limit int, filters map[string]interface{}) ([]model.EmailLog, int64, error) {
	var logs []model.EmailLog
	var total int64

	query := r.db.Model(&model.EmailLog{})

	// Apply filters
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if provider, ok := filters["provider"]; ok {
		query = query.Where("provider = ?", provider)
	}
	if fromDate, ok := filters["from_date"]; ok {
		query = query.Where("sent_at >= ?", fromDate)
	}
	if toDate, ok := filters["to_date"]; ok {
		query = query.Where("sent_at <= ?", toDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("sent_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *emailRepository) GetEmailStats() (*model.EmailStats, error) {
	stats := &model.EmailStats{}

	// Count sent emails
	r.db.Model(&model.EmailLog{}).Where("status = ?", model.EmailStatusSent).Count(&stats.TotalSent)
	r.db.Model(&model.EmailLog{}).Where("status = ?", model.EmailStatusFailed).Count(&stats.TotalFailed)
	r.db.Model(&model.EmailQueue{}).Where("status = ?", model.EmailStatusPending).Count(&stats.TotalPending)

	// Calculate success rate
	total := stats.TotalSent + stats.TotalFailed
	if total > 0 {
		stats.SuccessRate = float64(stats.TotalSent) / float64(total) * 100
	}

	// Get last sent time
	var lastSent model.EmailLog
	err := r.db.Where("status = ?", model.EmailStatusSent).Order("sent_at DESC").First(&lastSent).Error
	if err == nil {
		stats.LastSentAt = &lastSent.SentAt
	}

	return stats, nil
}

// Email Config
func (r *emailRepository) CreateEmailConfig(config *model.EmailConfig) error {
	return r.db.Create(config).Error
}

func (r *emailRepository) GetEmailConfigByID(id uint) (*model.EmailConfig, error) {
	var config model.EmailConfig
	err := r.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *emailRepository) GetActiveEmailConfig() (*model.EmailConfig, error) {
	var config model.EmailConfig
	err := r.db.Where("is_active = ?", true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *emailRepository) UpdateEmailConfig(config *model.EmailConfig) error {
	return r.db.Save(config).Error
}

func (r *emailRepository) DeleteEmailConfig(id uint) error {
	return r.db.Delete(&model.EmailConfig{}, id).Error
}

func (r *emailRepository) GetAllEmailConfigs() ([]model.EmailConfig, error) {
	var configs []model.EmailConfig
	err := r.db.Order("created_at DESC").Find(&configs).Error
	return configs, err
}
