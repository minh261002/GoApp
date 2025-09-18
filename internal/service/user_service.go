package service

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/email"
	"go_app/pkg/jwt"
	"go_app/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	otpRepo      repository.OTPRepository
	jwtManager   *jwt.JWTManager
	emailService *email.EmailService
}

type LoginRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	DeviceID  string `json:"device_id" validate:"required"`
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
}

type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name" validate:"max=50"`
	LastName  string `json:"last_name" validate:"max=50"`
	Phone     string `json:"phone" validate:"phone"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Code     string `json:"code" validate:"required,len=6"`
	Password string `json:"password" validate:"required,min=6"`
}

type AuthResponse struct {
	User         *model.User `json:"user"`
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    time.Time   `json:"expires_at"`
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo:     repository.NewUserRepository(),
		sessionRepo:  repository.NewSessionRepository(),
		otpRepo:      repository.NewOTPRepository(),
		jwtManager:   jwt.NewJWTManager(),
		emailService: email.NewEmailService(),
	}
}

// Register creates a new user account
func (s *AuthService) Register(req *RegisterRequest) (*model.User, error) {
	// Check if email already exists
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	existingUser, _ = s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      "user", // Default role
		IsActive:  true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send email verification OTP
	if err := s.sendEmailVerificationOTP(user); err != nil {
		logger.Warnf("Failed to send email verification OTP: %v", err)
	}

	return user, nil
}

// Login authenticates user and creates session
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if !s.verifyPassword(req.Password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Deactivate all existing sessions for this user (single device login)
	if err := s.sessionRepo.DeactivateAllUserSessions(user.ID); err != nil {
		logger.Warnf("Failed to deactivate existing sessions: %v", err)
	}

	// Create new session
	session, err := s.createSession(user.ID, req.DeviceID, req.UserAgent, req.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.Role,
		fmt.Sprintf("%d", session.ID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	if err := s.userRepo.Update(user); err != nil {
		logger.Warnf("Failed to update last login: %v", err)
	}

	return &AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// ForgotPassword sends OTP for password reset
func (s *AuthService) ForgotPassword(req *ForgotPasswordRequest) error {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// Generate and save OTP
	otpCode := s.generateOTPCode()
	otp := &model.OTP{
		UserID:    user.ID,
		Email:     user.Email,
		Code:      otpCode,
		Type:      model.OTPTypePasswordReset,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minutes expiry
	}

	if err := s.otpRepo.Create(otp); err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send OTP email
	if err := s.emailService.SendPasswordResetEmail(user.Email, otpCode); err != nil {
		logger.Warnf("Failed to send password reset email: %v", err)
	}

	return nil
}

// ResetPassword resets password using OTP
func (s *AuthService) ResetPassword(req *ResetPasswordRequest) error {
	// Get valid OTP
	otp, err := s.otpRepo.GetValidByCode(req.Code)
	if err != nil {
		return errors.New("invalid or expired OTP code")
	}

	// Get user
	user, err := s.userRepo.GetByID(otp.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.Password = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark OTP as used
	if err := s.otpRepo.MarkAsUsed(req.Code); err != nil {
		logger.Warnf("Failed to mark OTP as used: %v", err)
	}

	// Deactivate all sessions for security
	if err := s.sessionRepo.DeactivateAllUserSessions(user.ID); err != nil {
		logger.Warnf("Failed to deactivate sessions: %v", err)
	}

	return nil
}

// Logout deactivates user session
func (s *AuthService) Logout(sessionID string) error {
	// Get session by token (assuming sessionID is the token)
	session, err := s.sessionRepo.GetByToken(sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	// Deactivate session
	session.IsActive = false
	if err := s.sessionRepo.Update(session); err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	return nil
}

// VerifyEmail verifies user email using OTP
func (s *AuthService) VerifyEmail(code string) error {
	// Get valid OTP
	otp, err := s.otpRepo.GetValidByCode(code)
	if err != nil {
		return errors.New("invalid or expired OTP code")
	}

	if otp.Type != model.OTPTypeEmailVerify {
		return errors.New("invalid OTP type")
	}

	// Get user
	user, err := s.userRepo.GetByID(otp.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Mark email as verified
	user.IsEmailVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Mark OTP as used
	if err := s.otpRepo.MarkAsUsed(code); err != nil {
		logger.Warnf("Failed to mark OTP as used: %v", err)
	}

	return nil
}

// Helper methods
func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *AuthService) generateOTPCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *AuthService) createSession(userID uint, deviceID, userAgent, ipAddress string) (*model.Session, error) {
	// Generate session token
	sessionToken := uuid.New().String()

	// Calculate expiry time
	expireHours := 24
	if hours := os.Getenv("JWT_EXPIRE_HOURS"); hours != "" {
		if h, err := strconv.Atoi(hours); err == nil {
			expireHours = h
		}
	}
	expiresAt := time.Now().Add(time.Duration(expireHours) * time.Hour)

	session := &model.Session{
		UserID:    userID,
		Token:     sessionToken,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) sendEmailVerificationOTP(user *model.User) error {
	// Generate OTP
	otpCode := s.generateOTPCode()
	otp := &model.OTP{
		UserID:    user.ID,
		Email:     user.Email,
		Code:      otpCode,
		Type:      model.OTPTypeEmailVerify,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minutes expiry
	}

	if err := s.otpRepo.Create(otp); err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send email
	return s.emailService.SendEmailVerification(user.Email, otpCode)
}

// GetUserByID gets user by ID
func (s *AuthService) GetUserByID(userID uint) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}
