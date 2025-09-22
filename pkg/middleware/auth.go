package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"go_app/internal/repository"
	"go_app/pkg/jwt"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtManager  *jwt.JWTManager
	sessionRepo repository.SessionRepository
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:  jwt.NewJWTManager(),
		sessionRepo: repository.NewSessionRepository(),
	}
}

// AuthMiddleware validates JWT token and session
func (m *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate JWT token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}

		// Get session ID from claims
		sessionID, err := strconv.ParseUint(claims.SessionID, 10, 32)
		if err != nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "Invalid session")
			c.Abort()
			return
		}

		// Check if session exists and is active
		session, err := m.sessionRepo.GetByToken(token)
		if err != nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "Session not found")
			c.Abort()
			return
		}

		// Verify session ID matches
		if session.ID != uint(sessionID) {
			response.ErrorResponse(c, http.StatusUnauthorized, "Invalid session")
			c.Abort()
			return
		}

		// Check if session is valid (active and not expired)
		if !session.IsValid() {
			response.ErrorResponse(c, http.StatusUnauthorized, "Session expired or inactive")
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("session_id", session.ID)
		c.Set("session_token", token)

		c.Next()
	}
}

// AdminMiddleware checks if user is admin
func (m *AuthMiddleware) AdminMiddleware() gin.HandlerFunc {
	roleHelper := NewRoleHelper()
	return roleHelper.RequireRole("admin")
}

// ModeratorMiddleware checks if user is moderator or admin
func (m *AuthMiddleware) ModeratorMiddleware() gin.HandlerFunc {
	roleHelper := NewRoleHelper()
	return roleHelper.RequireAnyRole("admin", "moderator")
}

// OptionalAuthMiddleware validates JWT token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]

		// Validate JWT token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Get session ID from claims
		sessionID, err := strconv.ParseUint(claims.SessionID, 10, 32)
		if err != nil {
			c.Next()
			return
		}

		// Check if session exists and is active
		session, err := m.sessionRepo.GetByToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Verify session ID matches
		if session.ID != uint(sessionID) {
			c.Next()
			return
		}

		// Check if session is valid (active and not expired)
		if !session.IsValid() {
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("session_id", session.ID)
		c.Set("session_token", token)

		c.Next()
	}
}
