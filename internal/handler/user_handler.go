package handler

import (
	"net/http"

	"go_app/internal/service"
	"go_app/pkg/response"
	"go_app/pkg/validator"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Register request"
// @Success 201 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		response.ErrorResponse(c, http.StatusConflict, err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "User registered successfully", user)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and create session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login request"
// @Success 200 {object} response.Response{data=service.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get client IP and User-Agent
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		response.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Login successful", authResponse)
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send OTP to user email for password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req service.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.authService.ForgotPassword(&req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to process request")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "If the email exists, a password reset code has been sent", nil)
}

// ResetPassword godoc
// @Summary Reset password with OTP
// @Description Reset user password using OTP code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.authService.ResetPassword(&req)
	if err != nil {
		response.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Password reset successfully", nil)
}

// VerifyEmail godoc
// @Summary Verify email with OTP
// @Description Verify user email using OTP code
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "OTP code"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "OTP code is required")
		return
	}

	err := h.authService.VerifyEmail(code)
	if err != nil {
		response.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Email verified successfully", nil)
}

// Logout godoc
// @Summary User logout
// @Description Logout user and deactivate session
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionToken, exists := c.Get("session_token")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "Session not found")
		return
	}

	err := h.authService.Logout(sessionToken.(string))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to logout")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=model.User}
// @Failure 401 {object} response.Response
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not found")
		return
	}

	// Get user from repository
	userRepo := service.NewAuthService()
	user, err := userRepo.GetUserByID(userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", user)
}
