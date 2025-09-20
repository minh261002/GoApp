package handler

import (
	"net/http"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	permissionService service.PermissionService
}

// NewPermissionHandler creates a new PermissionHandler
func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{
		permissionService: service.NewPermissionService(),
	}
}

// Permissions

// CreatePermission creates a new permission
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req model.PermissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	permission, err := h.permissionService.CreatePermission(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create permission", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Permission created successfully", permission)
}

// GetPermissionByID retrieves a permission by its ID
func (h *PermissionHandler) GetPermissionByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID", err.Error())
		return
	}

	permission, err := h.permissionService.GetPermissionByID(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permission", err.Error())
		return
	}

	if permission == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Permission not found", "permission not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission retrieved successfully", permission)
}

// GetPermissionByName retrieves a permission by its name
func (h *PermissionHandler) GetPermissionByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Permission name is required", "name parameter is required")
		return
	}

	permission, err := h.permissionService.GetPermissionByName(name)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permission", err.Error())
		return
	}

	if permission == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Permission not found", "permission not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission retrieved successfully", permission)
}

// GetAllPermissions retrieves all permissions with pagination and filters
func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if resource := c.Query("resource"); resource != "" {
		filters["resource"] = resource
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active"] = active
		}
	}
	if isSystem := c.Query("is_system"); isSystem != "" {
		if system, err := strconv.ParseBool(isSystem); err == nil {
			filters["is_system"] = system
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	permissions, total, err := h.permissionService.GetAllPermissions(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permissions", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Permissions retrieved successfully", permissions, page, limit, total)
}

// UpdatePermission updates an existing permission
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID", err.Error())
		return
	}

	var req model.PermissionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	permission, err := h.permissionService.UpdatePermission(uint(id), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update permission", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission updated successfully", permission)
}

// DeletePermission deletes a permission
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.permissionService.DeletePermission(uint(id), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete permission", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission deleted successfully", nil)
}

// GetPermissionsByResource retrieves permissions by resource
func (h *PermissionHandler) GetPermissionsByResource(c *gin.Context) {
	resourceStr := c.Param("resource")
	if resourceStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Resource is required", "resource parameter is required")
		return
	}

	resource := model.ResourceType(resourceStr)
	permissions, err := h.permissionService.GetPermissionsByResource(resource)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permissions", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permissions retrieved successfully", permissions)
}

// Roles

// CreateRole creates a new role
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req model.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	role, err := h.permissionService.CreateRole(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create role", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Role created successfully", role)
}

// GetRoleByID retrieves a role by its ID
func (h *PermissionHandler) GetRoleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
		return
	}

	role, err := h.permissionService.GetRoleByID(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve role", err.Error())
		return
	}

	if role == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Role not found", "role not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Role retrieved successfully", role)
}

// GetRoleByName retrieves a role by its name
func (h *PermissionHandler) GetRoleByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Role name is required", "name parameter is required")
		return
	}

	role, err := h.permissionService.GetRoleByName(name)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve role", err.Error())
		return
	}

	if role == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Role not found", "role not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Role retrieved successfully", role)
}

// GetAllRoles retrieves all roles with pagination and filters
func (h *PermissionHandler) GetAllRoles(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active"] = active
		}
	}
	if isSystem := c.Query("is_system"); isSystem != "" {
		if system, err := strconv.ParseBool(isSystem); err == nil {
			filters["is_system"] = system
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	roles, total, err := h.permissionService.GetAllRoles(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve roles", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Roles retrieved successfully", roles, page, limit, total)
}

// UpdateRole updates an existing role
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
		return
	}

	var req model.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	role, err := h.permissionService.UpdateRole(uint(id), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update role", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Role updated successfully", role)
}

// DeleteRole deletes a role
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.permissionService.DeleteRole(uint(id), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete role", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Role deleted successfully", nil)
}

// Role Permissions

// AssignPermissionToRole assigns a permission to a role
func (h *PermissionHandler) AssignPermissionToRole(c *gin.Context) {
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.permissionService.AssignPermissionToRole(uint(roleID), uint(permissionID), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign permission to role", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission assigned to role successfully", nil)
}

// RevokePermissionFromRole revokes a permission from a role
func (h *PermissionHandler) RevokePermissionFromRole(c *gin.Context) {
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.permissionService.RevokePermissionFromRole(uint(roleID), uint(permissionID), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to revoke permission from role", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission revoked from role successfully", nil)
}

// GetRolePermissions retrieves permissions for a role
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
		return
	}

	permissions, err := h.permissionService.GetRolePermissions(uint(roleID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve role permissions", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Role permissions retrieved successfully", permissions)
}

// UpdateRolePermissions updates permissions for a role
func (h *PermissionHandler) UpdateRolePermissions(c *gin.Context) {
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
		return
	}

	var req model.RolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.permissionService.UpdateRolePermissions(uint(roleID), &req, userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update role permissions", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Role permissions updated successfully", nil)
}

// User Permissions

// AssignPermissionToUser assigns a permission to a user
func (h *PermissionHandler) AssignPermissionToUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req model.UserPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get granted by user ID from context
	grantedBy, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.permissionService.AssignPermissionToUser(uint(userID), req.PermissionID, grantedBy.(uint), &req); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign permission to user", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission assigned to user successfully", nil)
}

// RevokePermissionFromUser revokes a permission from a user
func (h *PermissionHandler) RevokePermissionFromUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID", err.Error())
		return
	}

	// Get granted by user ID from context
	grantedBy, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.permissionService.RevokePermissionFromUser(uint(userID), uint(permissionID), grantedBy.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to revoke permission from user", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission revoked from user successfully", nil)
}

// GetUserPermissions retrieves permissions for a user
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	permissions, err := h.permissionService.GetUserPermissions(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user permissions", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User permissions retrieved successfully", permissions)
}

// GetUserEffectivePermissions retrieves all effective permissions for a user
func (h *PermissionHandler) GetUserEffectivePermissions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	permissions, err := h.permissionService.GetUserEffectivePermissions(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user effective permissions", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User effective permissions retrieved successfully", permissions)
}

// Permission Checking

// CheckPermission checks if a user has permission for a specific resource and action
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req model.PermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	permissionCheck, err := h.permissionService.CheckPermission(req.UserID, req.Resource, req.Action, req.ResourceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check permission", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission check completed", permissionCheck)
}

// GetUserPermissionsForResource retrieves user permissions for a specific resource
func (h *PermissionHandler) GetUserPermissionsForResource(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	resourceStr := c.Param("resource")
	if resourceStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Resource is required", "resource parameter is required")
		return
	}

	resource := model.ResourceType(resourceStr)
	permissions, err := h.permissionService.GetUserPermissionsForResource(uint(userID), resource)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user permissions for resource", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User permissions for resource retrieved successfully", permissions)
}

// Audit & Logging

// GetPermissionLogs retrieves permission logs
func (h *PermissionHandler) GetPermissionLogs(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = uint(id)
		}
	}
	if targetUserID := c.Query("target_user_id"); targetUserID != "" {
		if id, err := strconv.ParseUint(targetUserID, 10, 32); err == nil {
			filters["target_user_id"] = uint(id)
		}
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if date, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["date_from"] = date
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if date, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["date_to"] = date
		}
	}

	logs, total, err := h.permissionService.GetPermissionLogs(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permission logs", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Permission logs retrieved successfully", logs, page, limit, total)
}

// Statistics

// GetPermissionStats retrieves permission statistics
func (h *PermissionHandler) GetPermissionStats(c *gin.Context) {
	stats, err := h.permissionService.GetPermissionStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permission statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Permission statistics retrieved successfully", stats)
}

// GetUserPermissionStats retrieves permission statistics for a specific user
func (h *PermissionHandler) GetUserPermissionStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	stats, err := h.permissionService.GetUserPermissionStats(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user permission statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User permission statistics retrieved successfully", stats)
}

// Utility

// InitializeDefaultPermissions initializes default permissions and roles
func (h *PermissionHandler) InitializeDefaultPermissions(c *gin.Context) {
	if err := h.permissionService.InitializeDefaultPermissions(); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to initialize default permissions", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Default permissions initialized successfully", nil)
}

// SyncUserRole synchronizes user role with permission system
func (h *PermissionHandler) SyncUserRole(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	roleName := c.Param("role_name")
	if roleName == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Role name is required", "role_name parameter is required")
		return
	}

	if err := h.permissionService.SyncUserRole(uint(userID), roleName); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to sync user role", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User role synchronized successfully", nil)
}
