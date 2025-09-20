package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"html/template"
	"net/smtp"
	"strings"
	"time"
)

type EmailService interface {
	// Email Templates
	CreateEmailTemplate(req *model.EmailTemplate) (*model.EmailTemplate, error)
	GetEmailTemplateByID(id uint) (*model.EmailTemplate, error)
	GetEmailTemplateByName(name string) (*model.EmailTemplate, error)
	GetEmailTemplatesByType(templateType string) ([]model.EmailTemplate, error)
	UpdateEmailTemplate(id uint, req *model.EmailTemplate) (*model.EmailTemplate, error)
	DeleteEmailTemplate(id uint) error
	GetAllEmailTemplates(page, limit int) ([]model.EmailTemplate, int64, error)

	// Email Sending
	SendEmail(req *model.EmailRequest) (*model.EmailResponse, error)
	SendEmailWithTemplate(templateName string, to []string, variables map[string]interface{}) (*model.EmailResponse, error)
	QueueEmail(req *model.EmailRequest) (*model.EmailResponse, error)
	ProcessEmailQueue() error
	RetryFailedEmails() error

	// Email Queue Management
	GetEmailQueueStats() (map[string]interface{}, error)
	GetEmailLogs(page, limit int, filters map[string]interface{}) ([]model.EmailLog, int64, error)
	GetEmailStats() (*model.EmailStats, error)

	// Email Config
	CreateEmailConfig(req *model.EmailConfig) (*model.EmailConfig, error)
	GetEmailConfigByID(id uint) (*model.EmailConfig, error)
	GetActiveEmailConfig() (*model.EmailConfig, error)
	UpdateEmailConfig(id uint, req *model.EmailConfig) (*model.EmailConfig, error)
	DeleteEmailConfig(id uint) error
	GetAllEmailConfigs() ([]model.EmailConfig, error)
}

type emailService struct {
	emailRepo repository.EmailRepository
}

func NewEmailService(emailRepo repository.EmailRepository) EmailService {
	return &emailService{
		emailRepo: emailRepo,
	}
}

// Email Templates
func (s *emailService) CreateEmailTemplate(req *model.EmailTemplate) (*model.EmailTemplate, error) {
	if err := s.emailRepo.CreateEmailTemplate(req); err != nil {
		logger.Errorf("Failed to create email template: %v", err)
		return nil, fmt.Errorf("failed to create email template")
	}
	return req, nil
}

func (s *emailService) GetEmailTemplateByID(id uint) (*model.EmailTemplate, error) {
	template, err := s.emailRepo.GetEmailTemplateByID(id)
	if err != nil {
		logger.Errorf("Failed to get email template by ID %d: %v", id, err)
		return nil, fmt.Errorf("email template not found")
	}
	return template, nil
}

func (s *emailService) GetEmailTemplateByName(name string) (*model.EmailTemplate, error) {
	template, err := s.emailRepo.GetEmailTemplateByName(name)
	if err != nil {
		logger.Errorf("Failed to get email template by name %s: %v", name, err)
		return nil, fmt.Errorf("email template not found")
	}
	return template, nil
}

func (s *emailService) GetEmailTemplatesByType(templateType string) ([]model.EmailTemplate, error) {
	templates, err := s.emailRepo.GetEmailTemplatesByType(templateType)
	if err != nil {
		logger.Errorf("Failed to get email templates by type %s: %v", templateType, err)
		return nil, fmt.Errorf("failed to get email templates")
	}
	return templates, nil
}

func (s *emailService) UpdateEmailTemplate(id uint, req *model.EmailTemplate) (*model.EmailTemplate, error) {
	req.ID = id
	if err := s.emailRepo.UpdateEmailTemplate(req); err != nil {
		logger.Errorf("Failed to update email template %d: %v", id, err)
		return nil, fmt.Errorf("failed to update email template")
	}
	return req, nil
}

func (s *emailService) DeleteEmailTemplate(id uint) error {
	if err := s.emailRepo.DeleteEmailTemplate(id); err != nil {
		logger.Errorf("Failed to delete email template %d: %v", id, err)
		return fmt.Errorf("failed to delete email template")
	}
	return nil
}

func (s *emailService) GetAllEmailTemplates(page, limit int) ([]model.EmailTemplate, int64, error) {
	templates, total, err := s.emailRepo.GetAllEmailTemplates(page, limit)
	if err != nil {
		logger.Errorf("Failed to get all email templates: %v", err)
		return nil, 0, fmt.Errorf("failed to get email templates")
	}
	return templates, total, nil
}

// Email Sending
func (s *emailService) SendEmail(req *model.EmailRequest) (*model.EmailResponse, error) {
	// Get active email config
	config, err := s.emailRepo.GetActiveEmailConfig()
	if err != nil {
		logger.Errorf("Failed to get active email config: %v", err)
		return nil, fmt.Errorf("email service not configured")
	}

	// Create email queue entry
	emailQueue := &model.EmailQueue{
		To:          strings.Join(req.To, ","),
		CC:          strings.Join(req.CC, ","),
		BCC:         strings.Join(req.BCC, ","),
		Subject:     req.Subject,
		Body:        req.Body,
		BodyHTML:    req.BodyHTML,
		Priority:    req.Priority,
		Status:      model.EmailStatusPending,
		ScheduledAt: req.ScheduledAt,
		MaxAttempts: 3,
	}

	if err := s.emailRepo.CreateEmailQueue(emailQueue); err != nil {
		logger.Errorf("Failed to create email queue: %v", err)
		return nil, fmt.Errorf("failed to queue email")
	}

	// Send immediately if not scheduled
	if req.ScheduledAt == nil || req.ScheduledAt.Before(time.Now()) {
		if err := s.sendEmailNow(emailQueue, config); err != nil {
			logger.Errorf("Failed to send email immediately: %v", err)
			return &model.EmailResponse{
				ID:        emailQueue.ID,
				Status:    model.EmailStatusFailed,
				Message:   err.Error(),
				CreatedAt: emailQueue.CreatedAt,
			}, nil
		}
	}

	return &model.EmailResponse{
		ID:        emailQueue.ID,
		Status:    emailQueue.Status,
		Message:   "Email queued successfully",
		CreatedAt: emailQueue.CreatedAt,
	}, nil
}

func (s *emailService) SendEmailWithTemplate(templateName string, to []string, variables map[string]interface{}) (*model.EmailResponse, error) {
	// Get template
	template, err := s.emailRepo.GetEmailTemplateByName(templateName)
	if err != nil {
		logger.Errorf("Failed to get email template %s: %v", templateName, err)
		return nil, fmt.Errorf("email template not found")
	}

	// Process template
	subject, err := s.processTemplate(template.Subject, variables)
	if err != nil {
		logger.Errorf("Failed to process subject template: %v", err)
		return nil, fmt.Errorf("failed to process email template")
	}

	body, err := s.processTemplate(template.Body, variables)
	if err != nil {
		logger.Errorf("Failed to process body template: %v", err)
		return nil, fmt.Errorf("failed to process email template")
	}

	var bodyHTML string
	if template.BodyHTML != "" {
		bodyHTML, err = s.processTemplate(template.BodyHTML, variables)
		if err != nil {
			logger.Errorf("Failed to process HTML body template: %v", err)
			return nil, fmt.Errorf("failed to process email template")
		}
	}

	// Create email request
	emailReq := &model.EmailRequest{
		To:       to,
		Subject:  subject,
		Body:     body,
		BodyHTML: bodyHTML,
		Priority: model.EmailPriorityNormal,
	}

	return s.SendEmail(emailReq)
}

func (s *emailService) QueueEmail(req *model.EmailRequest) (*model.EmailResponse, error) {
	return s.SendEmail(req)
}

func (s *emailService) ProcessEmailQueue() error {
	// Get pending emails
	emails, err := s.emailRepo.GetPendingEmails(10) // Process 10 emails at a time
	if err != nil {
		logger.Errorf("Failed to get pending emails: %v", err)
		return fmt.Errorf("failed to get pending emails")
	}

	if len(emails) == 0 {
		return nil // No emails to process
	}

	// Get active email config
	config, err := s.emailRepo.GetActiveEmailConfig()
	if err != nil {
		logger.Errorf("Failed to get active email config: %v", err)
		return fmt.Errorf("email service not configured")
	}

	// Process each email
	for _, email := range emails {
		if err := s.sendEmailNow(&email, config); err != nil {
			logger.Errorf("Failed to send email %d: %v", email.ID, err)
			// Continue with next email
		}
	}

	// Also retry failed emails as part of queue processing
	if err := s.RetryFailedEmails(); err != nil {
		logger.Errorf("Failed to retry failed emails: %v", err)
		// Don't return error here as it's not critical
	}

	return nil
}

func (s *emailService) RetryFailedEmails() error {
	// Get failed emails that haven't exceeded max attempts
	failedEmails, err := s.emailRepo.GetFailedEmails(10) // Retry 10 emails at a time
	if err != nil {
		logger.Errorf("Failed to get failed emails: %v", err)
		return fmt.Errorf("failed to get failed emails")
	}

	if len(failedEmails) == 0 {
		return nil // No failed emails to retry
	}

	// Get active email config
	config, err := s.emailRepo.GetActiveEmailConfig()
	if err != nil {
		logger.Errorf("Failed to get active email config: %v", err)
		return fmt.Errorf("email service not configured")
	}

	// Retry each failed email
	for _, email := range failedEmails {
		// Reset status to pending for retry
		email.Status = model.EmailStatusPending
		email.ErrorMessage = ""

		// Update in database
		if err := s.emailRepo.UpdateEmailQueue(&email); err != nil {
			logger.Errorf("Failed to update email %d for retry: %v", email.ID, err)
			continue
		}

		// Try to send the email
		if err := s.sendEmailNow(&email, config); err != nil {
			logger.Errorf("Failed to retry email %d: %v", email.ID, err)
			// Continue with next email
		} else {
			logger.Infof("Successfully retried email %d", email.ID)
		}
	}

	return nil
}

// Helper methods
func (s *emailService) sendEmailNow(email *model.EmailQueue, config *model.EmailConfig) error {
	// Update status to sending and increment attempts
	email.Status = model.EmailStatusSending
	email.Attempts++
	if err := s.emailRepo.UpdateEmailQueue(email); err != nil {
		logger.Errorf("Failed to update email status to sending: %v", err)
	}

	// Send email via SMTP
	if err := s.sendSMTPEmail(email, config); err != nil {
		// Update status to failed
		email.Status = model.EmailStatusFailed
		email.ErrorMessage = err.Error()
		s.emailRepo.UpdateEmailQueue(email)

		// Create log entry
		log := &model.EmailLog{
			EmailQueueID: email.ID,
			To:           email.To,
			Subject:      email.Subject,
			Status:       model.EmailStatusFailed,
			ErrorMessage: err.Error(),
			Provider:     config.Provider,
			SentAt:       time.Now(),
		}
		s.emailRepo.CreateEmailLog(log)

		return err
	}

	// Update status to sent
	now := time.Now()
	email.Status = model.EmailStatusSent
	email.SentAt = &now
	if err := s.emailRepo.UpdateEmailQueue(email); err != nil {
		logger.Errorf("Failed to update email status to sent: %v", err)
	}

	// Create log entry
	log := &model.EmailLog{
		EmailQueueID: email.ID,
		To:           email.To,
		Subject:      email.Subject,
		Status:       model.EmailStatusSent,
		Provider:     config.Provider,
		SentAt:       now,
	}
	s.emailRepo.CreateEmailLog(log)

	return nil
}

func (s *emailService) sendSMTPEmail(email *model.EmailQueue, config *model.EmailConfig) error {
	// Setup authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// Create message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", config.FromName, config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", email.To))
	if email.CC != "" {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", email.CC))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if email.BodyHTML != "" {
		msg.WriteString("Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n")
		msg.WriteString("\r\n--boundary123\r\n")
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(email.Body)
		msg.WriteString("\r\n--boundary123\r\n")
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		msg.WriteString(email.BodyHTML)
		msg.WriteString("\r\n--boundary123--\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(email.Body)
	}

	// Send email
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	to := strings.Split(email.To, ",")
	if email.CC != "" {
		to = append(to, strings.Split(email.CC, ",")...)
	}
	if email.BCC != "" {
		to = append(to, strings.Split(email.BCC, ",")...)
	}

	if config.TLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: !config.SSL,
			ServerName:         config.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, config.Host)
		if err != nil {
			return err
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return err
		}

		if err = client.Mail(config.From); err != nil {
			return err
		}

		for _, addr := range to {
			if err = client.Rcpt(addr); err != nil {
				return err
			}
		}

		w, err := client.Data()
		if err != nil {
			return err
		}
		defer w.Close()

		_, err = w.Write(msg.Bytes())
		return err
	} else {
		return smtp.SendMail(addr, auth, config.From, to, msg.Bytes())
	}
}

func (s *emailService) processTemplate(templateStr string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Email Queue Management
func (s *emailService) GetEmailQueueStats() (map[string]interface{}, error) {
	return s.emailRepo.GetEmailQueueStats()
}

func (s *emailService) GetEmailLogs(page, limit int, filters map[string]interface{}) ([]model.EmailLog, int64, error) {
	return s.emailRepo.GetEmailLogs(page, limit, filters)
}

func (s *emailService) GetEmailStats() (*model.EmailStats, error) {
	return s.emailRepo.GetEmailStats()
}

// Email Config
func (s *emailService) CreateEmailConfig(req *model.EmailConfig) (*model.EmailConfig, error) {
	if err := s.emailRepo.CreateEmailConfig(req); err != nil {
		logger.Errorf("Failed to create email config: %v", err)
		return nil, fmt.Errorf("failed to create email config")
	}
	return req, nil
}

func (s *emailService) GetEmailConfigByID(id uint) (*model.EmailConfig, error) {
	config, err := s.emailRepo.GetEmailConfigByID(id)
	if err != nil {
		logger.Errorf("Failed to get email config by ID %d: %v", id, err)
		return nil, fmt.Errorf("email config not found")
	}
	return config, nil
}

func (s *emailService) GetActiveEmailConfig() (*model.EmailConfig, error) {
	config, err := s.emailRepo.GetActiveEmailConfig()
	if err != nil {
		logger.Errorf("Failed to get active email config: %v", err)
		return nil, fmt.Errorf("no active email config found")
	}
	return config, nil
}

func (s *emailService) UpdateEmailConfig(id uint, req *model.EmailConfig) (*model.EmailConfig, error) {
	req.ID = id
	if err := s.emailRepo.UpdateEmailConfig(req); err != nil {
		logger.Errorf("Failed to update email config %d: %v", id, err)
		return nil, fmt.Errorf("failed to update email config")
	}
	return req, nil
}

func (s *emailService) DeleteEmailConfig(id uint) error {
	if err := s.emailRepo.DeleteEmailConfig(id); err != nil {
		logger.Errorf("Failed to delete email config %d: %v", id, err)
		return fmt.Errorf("failed to delete email config")
	}
	return nil
}

func (s *emailService) GetAllEmailConfigs() ([]model.EmailConfig, error) {
	configs, err := s.emailRepo.GetAllEmailConfigs()
	if err != nil {
		logger.Errorf("Failed to get all email configs: %v", err)
		return nil, fmt.Errorf("failed to get email configs")
	}
	return configs, nil
}
