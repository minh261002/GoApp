package email

import (
	"fmt"
	"os"
	"strconv"

	"go_app/pkg/logger"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func NewEmailService() *EmailService {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &EmailService{
		smtpHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		smtpPort:     port,
		smtpUsername: getEnv("SMTP_USERNAME", ""),
		smtpPassword: getEnv("SMTP_PASSWORD", ""),
		fromEmail:    getEnv("FROM_EMAIL", ""),
		fromName:     getEnv("FROM_NAME", "Go App"),
	}
}

func (e *EmailService) SendOTPEmail(to, otpCode, otpType string) error {
	subject := "Mã xác thực OTP"
	body := e.generateOTPEmailBody(otpCode, otpType)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) SendPasswordResetEmail(to, otpCode string) error {
	subject := "Đặt lại mật khẩu"
	body := e.generatePasswordResetEmailBody(otpCode)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) SendEmailVerification(to, otpCode string) error {
	subject := "Xác thực email"
	body := e.generateEmailVerificationBody(otpCode)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) sendEmail(to, subject, body string) error {
	if e.smtpUsername == "" || e.smtpPassword == "" || e.fromEmail == "" {
		logger.Warn("Email configuration not set, skipping email send")
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", e.fromName, e.fromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.smtpHost, e.smtpPort, e.smtpUsername, e.smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		logger.Error("Failed to send email: %v", err)
		return err
	}

	logger.Info("Email sent successfully to: %s", to)
	return nil
}

func (e *EmailService) generateOTPEmailBody(otpCode, otpType string) string {
	var action string
	switch otpType {
	case "password_reset":
		action = "đặt lại mật khẩu"
	case "email_verify":
		action = "xác thực email"
	default:
		action = "xác thực"
	}

	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Mã xác thực OTP</title>
		</head>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
			<div style="background-color: #f8f9fa; padding: 30px; border-radius: 10px;">
				<h2 style="color: #333; text-align: center;">Mã xác thực OTP</h2>
				<p style="color: #666; font-size: 16px;">Xin chào,</p>
				<p style="color: #666; font-size: 16px;">Bạn đã yêu cầu %s. Vui lòng sử dụng mã OTP sau để hoàn tất quá trình:</p>
				<div style="background-color: #007bff; color: white; padding: 20px; text-align: center; border-radius: 5px; margin: 20px 0;">
					<h1 style="margin: 0; font-size: 32px; letter-spacing: 5px;">%s</h1>
				</div>
				<p style="color: #666; font-size: 14px;">Mã này có hiệu lực trong 10 phút. Vui lòng không chia sẻ mã này với bất kỳ ai.</p>
				<p style="color: #666; font-size: 14px;">Nếu bạn không yêu cầu %s, vui lòng bỏ qua email này.</p>
				<hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
				<p style="color: #999; font-size: 12px; text-align: center;">Email này được gửi tự động, vui lòng không trả lời.</p>
			</div>
		</body>
		</html>
	`, action, otpCode, action)
}

func (e *EmailService) generatePasswordResetEmailBody(otpCode string) string {
	return e.generateOTPEmailBody(otpCode, "password_reset")
}

func (e *EmailService) generateEmailVerificationBody(otpCode string) string {
	return e.generateOTPEmailBody(otpCode, "email_verify")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
