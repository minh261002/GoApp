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

type EmailHandler struct {
	emailService service.EmailService
}

func NewEmailHandler(emailService service.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// Email Templates
// CreateEmailTemplate creates a new email template
// @Summary Create email template
// @Description Create a new email template
// @Tags email-templates
// @Accept json
// @Produce json
// @Param template body model.EmailTemplate true "Email template data"
// @Success 201 {object} response.Response{data=model.EmailTemplate}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/templates [post]
func (h *EmailHandler) CreateEmailTemplate(c *gin.Context) {
	var req model.EmailTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	template, err := h.emailService.CreateEmailTemplate(&req)
	if err != nil {
		logger.Errorf("Failed to create email template: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create email template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Email template created successfully", template)
}

// GetEmailTemplateByID gets an email template by ID
// @Summary Get email template by ID
// @Description Get an email template by its ID
// @Tags email-templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} response.Response{data=model.EmailTemplate}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/templates/{id} [get]
func (h *EmailHandler) GetEmailTemplateByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	template, err := h.emailService.GetEmailTemplateByID(uint(id))
	if err != nil {
		if err.Error() == "email template not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Email template not found", nil)
			return
		}
		logger.Errorf("Failed to get email template: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email template retrieved successfully", template)
}

// GetEmailTemplatesByType gets email templates by type
// @Summary Get email templates by type
// @Description Get email templates filtered by type
// @Tags email-templates
// @Produce json
// @Param type query string true "Template type"
// @Success 200 {object} response.Response{data=[]model.EmailTemplate}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/templates/type/{type} [get]
func (h *EmailHandler) GetEmailTemplatesByType(c *gin.Context) {
	templateType := c.Param("type")
	if templateType == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Template type is required", nil)
		return
	}

	templates, err := h.emailService.GetEmailTemplatesByType(templateType)
	if err != nil {
		logger.Errorf("Failed to get email templates by type: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email templates", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email templates retrieved successfully", templates)
}

// UpdateEmailTemplate updates an email template
// @Summary Update email template
// @Description Update an existing email template
// @Tags email-templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param template body model.EmailTemplate true "Email template data"
// @Success 200 {object} response.Response{data=model.EmailTemplate}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/templates/{id} [put]
func (h *EmailHandler) UpdateEmailTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	var req model.EmailTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	template, err := h.emailService.UpdateEmailTemplate(uint(id), &req)
	if err != nil {
		if err.Error() == "email template not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Email template not found", nil)
			return
		}
		logger.Errorf("Failed to update email template: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update email template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email template updated successfully", template)
}

// DeleteEmailTemplate deletes an email template
// @Summary Delete email template
// @Description Delete an email template by ID
// @Tags email-templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/templates/{id} [delete]
func (h *EmailHandler) DeleteEmailTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	err = h.emailService.DeleteEmailTemplate(uint(id))
	if err != nil {
		if err.Error() == "email template not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Email template not found", nil)
			return
		}
		logger.Errorf("Failed to delete email template: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete email template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email template deleted successfully", nil)
}

// GetAllEmailTemplates gets all email templates with pagination
// @Summary Get all email templates
// @Description Get all email templates with pagination
// @Tags email-templates
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=[]model.EmailTemplate}
// @Failure 500 {object} response.Response
// @Router /api/v1/email/templates [get]
func (h *EmailHandler) GetAllEmailTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	templates, total, err := h.emailService.GetAllEmailTemplates(page, limit)
	if err != nil {
		logger.Errorf("Failed to get all email templates: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email templates", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email templates retrieved successfully", gin.H{
		"templates": templates,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// Email Sending
// SendEmail sends an email
// @Summary Send email
// @Description Send an email immediately or queue it for later
// @Tags emails
// @Accept json
// @Produce json
// @Param email body model.EmailRequest true "Email data"
// @Success 200 {object} response.Response{data=model.EmailResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/send [post]
func (h *EmailHandler) SendEmail(c *gin.Context) {
	var req model.EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	emailResp, err := h.emailService.SendEmail(&req)
	if err != nil {
		logger.Errorf("Failed to send email: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to send email", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email sent successfully", emailResp)
}

// SendEmailWithTemplate sends an email using a template
// @Summary Send email with template
// @Description Send an email using a predefined template
// @Tags emails
// @Accept json
// @Produce json
// @Param template_name path string true "Template name"
// @Param email body map[string]interface{} true "Email data with variables"
// @Success 200 {object} response.Response{data=model.EmailResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/send-template/{template_name} [post]
func (h *EmailHandler) SendEmailWithTemplate(c *gin.Context) {
	templateName := c.Param("template_name")
	if templateName == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Template name is required", nil)
		return
	}

	var req struct {
		To        []string               `json:"to" validate:"required,min=1"`
		Variables map[string]interface{} `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	emailResp, err := h.emailService.SendEmailWithTemplate(templateName, req.To, req.Variables)
	if err != nil {
		logger.Errorf("Failed to send email with template: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to send email", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email sent successfully", emailResp)
}

// ProcessEmailQueue processes pending emails
// @Summary Process email queue
// @Description Process pending emails in the queue
// @Tags emails
// @Produce json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/process-queue [post]
func (h *EmailHandler) ProcessEmailQueue(c *gin.Context) {
	err := h.emailService.ProcessEmailQueue()
	if err != nil {
		logger.Errorf("Failed to process email queue: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to process email queue", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email queue processed successfully", nil)
}

// RetryFailedEmails retries failed emails
// @Summary Retry failed emails
// @Description Retry sending failed emails that haven't exceeded max attempts
// @Tags emails
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/v1/email/retry-failed [post]
func (h *EmailHandler) RetryFailedEmails(c *gin.Context) {
	err := h.emailService.RetryFailedEmails()
	if err != nil {
		logger.Errorf("Failed to retry failed emails: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retry failed emails", err.Error())
		return
	}

	// Get updated queue stats
	stats, err := h.emailService.GetEmailQueueStats()
	if err != nil {
		logger.Errorf("Failed to get email queue stats: %v", err)
		response.SuccessResponse(c, http.StatusOK, "Failed emails retried successfully", nil)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Failed emails retried successfully", gin.H{
		"message": "Retry process completed",
		"stats":   stats,
	})
}

// GetEmailQueueStats gets email queue statistics
// @Summary Get email queue stats
// @Description Get statistics about the email queue
// @Tags emails
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/v1/email/queue-stats [get]
func (h *EmailHandler) GetEmailQueueStats(c *gin.Context) {
	stats, err := h.emailService.GetEmailQueueStats()
	if err != nil {
		logger.Errorf("Failed to get email queue stats: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email queue stats", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email queue stats retrieved successfully", stats)
}

// GetEmailLogs gets email logs
// @Summary Get email logs
// @Description Get email sending logs with pagination and filters
// @Tags emails
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param provider query string false "Filter by provider"
// @Success 200 {object} response.Response{data=[]model.EmailLog}
// @Failure 500 {object} response.Response
// @Router /api/v1/email/logs [get]
func (h *EmailHandler) GetEmailLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if provider := c.Query("provider"); provider != "" {
		filters["provider"] = provider
	}

	logs, total, err := h.emailService.GetEmailLogs(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get email logs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email logs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email logs retrieved successfully", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetEmailStats gets email statistics
// @Summary Get email stats
// @Description Get overall email statistics
// @Tags emails
// @Produce json
// @Success 200 {object} response.Response{data=model.EmailStats}
// @Failure 500 {object} response.Response
// @Router /api/v1/email/stats [get]
func (h *EmailHandler) GetEmailStats(c *gin.Context) {
	stats, err := h.emailService.GetEmailStats()
	if err != nil {
		logger.Errorf("Failed to get email stats: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email stats", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email stats retrieved successfully", stats)
}

// Email Config
// CreateEmailConfig creates a new email configuration
// @Summary Create email config
// @Description Create a new email configuration
// @Tags email-configs
// @Accept json
// @Produce json
// @Param config body model.EmailConfig true "Email config data"
// @Success 201 {object} response.Response{data=model.EmailConfig}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/configs [post]
func (h *EmailHandler) CreateEmailConfig(c *gin.Context) {
	var req model.EmailConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	config, err := h.emailService.CreateEmailConfig(&req)
	if err != nil {
		logger.Errorf("Failed to create email config: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create email config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Email config created successfully", config)
}

// GetEmailConfigByID gets an email configuration by ID
// @Summary Get email config by ID
// @Description Get an email configuration by its ID
// @Tags email-configs
// @Produce json
// @Param id path int true "Config ID"
// @Success 200 {object} response.Response{data=model.EmailConfig}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/configs/{id} [get]
func (h *EmailHandler) GetEmailConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID", err.Error())
		return
	}

	config, err := h.emailService.GetEmailConfigByID(uint(id))
	if err != nil {
		if err.Error() == "email config not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Email config not found", nil)
			return
		}
		logger.Errorf("Failed to get email config: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email config retrieved successfully", config)
}

// GetActiveEmailConfig gets the active email configuration
// @Summary Get active email config
// @Description Get the currently active email configuration
// @Tags email-configs
// @Produce json
// @Success 200 {object} response.Response{data=model.EmailConfig}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/configs/active [get]
func (h *EmailHandler) GetActiveEmailConfig(c *gin.Context) {
	config, err := h.emailService.GetActiveEmailConfig()
	if err != nil {
		if err.Error() == "no active email config found" {
			response.ErrorResponse(c, http.StatusNotFound, "No active email config found", nil)
			return
		}
		logger.Errorf("Failed to get active email config: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active email config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Active email config retrieved successfully", config)
}

// UpdateEmailConfig updates an email configuration
// @Summary Update email config
// @Description Update an existing email configuration
// @Tags email-configs
// @Accept json
// @Produce json
// @Param id path int true "Config ID"
// @Param config body model.EmailConfig true "Email config data"
// @Success 200 {object} response.Response{data=model.EmailConfig}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/configs/{id} [put]
func (h *EmailHandler) UpdateEmailConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID", err.Error())
		return
	}

	var req model.EmailConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	config, err := h.emailService.UpdateEmailConfig(uint(id), &req)
	if err != nil {
		if err.Error() == "email config not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Email config not found", nil)
			return
		}
		logger.Errorf("Failed to update email config: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update email config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email config updated successfully", config)
}

// DeleteEmailConfig deletes an email configuration
// @Summary Delete email config
// @Description Delete an email configuration by ID
// @Tags email-configs
// @Produce json
// @Param id path int true "Config ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/email/configs/{id} [delete]
func (h *EmailHandler) DeleteEmailConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID", err.Error())
		return
	}

	err = h.emailService.DeleteEmailConfig(uint(id))
	if err != nil {
		if err.Error() == "email config not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Email config not found", nil)
			return
		}
		logger.Errorf("Failed to delete email config: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete email config", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email config deleted successfully", nil)
}

// GetAllEmailConfigs gets all email configurations
// @Summary Get all email configs
// @Description Get all email configurations
// @Tags email-configs
// @Produce json
// @Success 200 {object} response.Response{data=[]model.EmailConfig}
// @Failure 500 {object} response.Response
// @Router /api/v1/email/configs [get]
func (h *EmailHandler) GetAllEmailConfigs(c *gin.Context) {
	configs, err := h.emailService.GetAllEmailConfigs()
	if err != nil {
		logger.Errorf("Failed to get all email configs: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get email configs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email configs retrieved successfully", configs)
}

// TestEmailRetry tests email retry functionality
// @Summary Test email retry
// @Description Test email retry functionality by creating a test email that will fail
// @Tags emails
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/v1/email/test-retry [post]
func (h *EmailHandler) TestEmailRetry(c *gin.Context) {
	// Create a test email that will fail (invalid SMTP config)
	testEmail := &model.EmailRequest{
		To:       []string{"test@example.com"},
		Subject:  "Test Email for Retry",
		Body:     "This is a test email to test retry functionality",
		Priority: model.EmailPriorityNormal,
	}

	// Send the email (it will likely fail due to invalid config)
	emailResp, err := h.emailService.SendEmail(testEmail)
	if err != nil {
		logger.Errorf("Failed to send test email: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to send test email", err.Error())
		return
	}

	// Get queue stats
	stats, err := h.emailService.GetEmailQueueStats()
	if err != nil {
		logger.Errorf("Failed to get email queue stats: %v", err)
		response.SuccessResponse(c, http.StatusOK, "Test email created", gin.H{
			"email_id": emailResp.ID,
			"message":  "Test email created successfully",
		})
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Test email created successfully", gin.H{
		"email_id": emailResp.ID,
		"message":  "Test email created. Check queue stats to see retry functionality.",
		"stats":    stats,
	})
}
