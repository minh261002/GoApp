package middleware

import (
	"net/http"

	"go_app/internal/repository"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// RoleHelper provides helper functions for role-based access control
type RoleHelper struct {
	userRepo repository.UserRepository
}

// NewRoleHelper creates a new RoleHelper instance
func NewRoleHelper() *RoleHelper {
	return &RoleHelper{
		userRepo: repository.NewUserRepository(),
	}
}

// CheckUserRole checks if the authenticated user has the specified role
func (h *RoleHelper) CheckUserRole(c *gin.Context, roleName string) bool {
	userID, exists := c.Get("user_id")
	if !exists {
		return false
	}

	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil || user == nil || user.UserRole == nil {
		return false
	}

	return user.UserRole.Name == roleName
}

// CheckUserAnyRole checks if the authenticated user has any of the specified roles
func (h *RoleHelper) CheckUserAnyRole(c *gin.Context, roleNames ...string) bool {
	userID, exists := c.Get("user_id")
	if !exists {
		return false
	}

	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil || user == nil || user.UserRole == nil {
		return false
	}

	for _, roleName := range roleNames {
		if user.UserRole.Name == roleName {
			return true
		}
	}

	return false
}

// RequireRole middleware that requires a specific role
func (h *RoleHelper) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		user, err := h.userRepo.GetByID(userID.(uint))
		if err != nil || user == nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not found")
			c.Abort()
			return
		}

		if user.UserRole == nil || user.UserRole.Name != roleName {
			response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "Required role: "+roleName)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (h *RoleHelper) RequireAnyRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		user, err := h.userRepo.GetByID(userID.(uint))
		if err != nil || user == nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not found")
			c.Abort()
			return
		}

		if user.UserRole == nil {
			response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "No role assigned")
			c.Abort()
			return
		}

		hasRole := false
		for _, roleName := range roleNames {
			if user.UserRole.Name == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "Required one of roles: "+joinStrings(roleNames, ", "))
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
