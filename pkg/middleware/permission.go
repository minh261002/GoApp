package middleware

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware checks if user has required permission
func PermissionMiddleware(resource model.ResourceType, action model.PermissionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
			c.Abort()
			return
		}

		// Get resource ID if provided in URL params
		var resourceID *uint
		if idStr := c.Param("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				resourceIDUint := uint(id)
				resourceID = &resourceIDUint
			}
		}

		// Check permission
		permissionService := service.NewPermissionService()
		permissionCheck, err := permissionService.CheckPermission(userID.(uint), resource, action, resourceID)
		if err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check permission", err.Error())
			c.Abort()
			return
		}

		if !permissionCheck.HasPermission {
			response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", permissionCheck.Reason)
			c.Abort()
			return
		}

		// Store permission info in context for potential use in handlers
		c.Set("permission_source", permissionCheck.Source)
		c.Set("permission_resource", resource)
		c.Set("permission_action", action)

		c.Next()
	}
}

// ResourcePermissionMiddleware checks permission for a specific resource
func ResourcePermissionMiddleware(resource model.ResourceType, action model.PermissionType, resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
			c.Abort()
			return
		}

		// Get resource ID from specified parameter
		var resourceID *uint
		if idStr := c.Param(resourceIDParam); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				resourceIDUint := uint(id)
				resourceID = &resourceIDUint
			}
		}

		// Check permission
		permissionService := service.NewPermissionService()
		permissionCheck, err := permissionService.CheckPermission(userID.(uint), resource, action, resourceID)
		if err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check permission", err.Error())
			c.Abort()
			return
		}

		if !permissionCheck.HasPermission {
			response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", permissionCheck.Reason)
			c.Abort()
			return
		}

		// Store permission info in context
		c.Set("permission_source", permissionCheck.Source)
		c.Set("permission_resource", resource)
		c.Set("permission_action", action)
		c.Set("resource_id", resourceID)

		c.Next()
	}
}

// MultiplePermissionMiddleware checks if user has any of the specified permissions
func MultiplePermissionMiddleware(permissions []PermissionRequirement) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
			c.Abort()
			return
		}

		permissionService := service.NewPermissionService()
		hasAnyPermission := false
		var lastError error

		for _, perm := range permissions {
			// Get resource ID if specified
			var resourceID *uint
			if perm.ResourceIDParam != "" {
				if idStr := c.Param(perm.ResourceIDParam); idStr != "" {
					if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
						resourceIDUint := uint(id)
						resourceID = &resourceIDUint
					}
				}
			}

			permissionCheck, err := permissionService.CheckPermission(userID.(uint), perm.Resource, perm.Action, resourceID)
			if err != nil {
				lastError = err
				continue
			}

			if permissionCheck.HasPermission {
				hasAnyPermission = true
				c.Set("permission_source", permissionCheck.Source)
				c.Set("permission_resource", perm.Resource)
				c.Set("permission_action", perm.Action)
				break
			}
		}

		if !hasAnyPermission {
			if lastError != nil {
				response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check permissions", lastError.Error())
			} else {
				response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "User does not have any of the required permissions")
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionRequirement represents a permission requirement
type PermissionRequirement struct {
	Resource        model.ResourceType
	Action          model.PermissionType
	ResourceIDParam string // Parameter name to get resource ID from
}

// AdminPermissionMiddleware checks if user has admin permission for a resource
func AdminPermissionMiddleware(resource model.ResourceType) gin.HandlerFunc {
	return PermissionMiddleware(resource, model.PermissionTypeAdmin)
}

// ManagePermissionMiddleware checks if user has manage permission for a resource
func ManagePermissionMiddleware(resource model.ResourceType) gin.HandlerFunc {
	return PermissionMiddleware(resource, model.PermissionTypeManage)
}

// WritePermissionMiddleware checks if user has write permission for a resource
func WritePermissionMiddleware(resource model.ResourceType) gin.HandlerFunc {
	return PermissionMiddleware(resource, model.PermissionTypeWrite)
}

// ReadPermissionMiddleware checks if user has read permission for a resource
func ReadPermissionMiddleware(resource model.ResourceType) gin.HandlerFunc {
	return PermissionMiddleware(resource, model.PermissionTypeRead)
}

// DeletePermissionMiddleware checks if user has delete permission for a resource
func DeletePermissionMiddleware(resource model.ResourceType) gin.HandlerFunc {
	return PermissionMiddleware(resource, model.PermissionTypeDelete)
}

// OwnershipPermissionMiddleware checks if user owns the resource or has admin permission
func OwnershipPermissionMiddleware(resource model.ResourceType, ownerIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
			c.Abort()
			return
		}

		// Get owner ID from parameter
		ownerIDStr := c.Param(ownerIDParam)
		if ownerIDStr == "" {
			response.ErrorResponse(c, http.StatusBadRequest, "Owner ID parameter missing", "owner_id parameter is required")
			c.Abort()
			return
		}

		ownerID, err := strconv.ParseUint(ownerIDStr, 10, 32)
		if err != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "Invalid owner ID", err.Error())
			c.Abort()
			return
		}

		// Check if user is the owner
		if userID.(uint) == uint(ownerID) {
			c.Set("permission_source", "ownership")
			c.Next()
			return
		}

		// Check if user has admin permission
		permissionService := service.NewPermissionService()
		permissionCheck, err := permissionService.CheckPermission(userID.(uint), resource, model.PermissionTypeAdmin, nil)
		if err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check permission", err.Error())
			c.Abort()
			return
		}

		if !permissionCheck.HasPermission {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", "You can only access your own resources or need admin permission")
			c.Abort()
			return
		}

		c.Set("permission_source", "admin")
		c.Next()
	}
}

// ConditionalPermissionMiddleware checks permission based on conditions
func ConditionalPermissionMiddleware(condition func(*gin.Context) bool, permission PermissionRequirement) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If condition is not met, skip permission check
		if !condition(c) {
			c.Next()
			return
		}

		// Apply permission check
		PermissionMiddleware(permission.Resource, permission.Action)(c)
	}
}

// ResourceTypePermissionMiddleware checks permission based on resource type from parameter
func ResourceTypePermissionMiddleware(resourceTypeParam string, action model.PermissionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
			c.Abort()
			return
		}

		// Get resource type from parameter
		resourceTypeStr := c.Param(resourceTypeParam)
		if resourceTypeStr == "" {
			response.ErrorResponse(c, http.StatusBadRequest, "Resource type parameter missing", "resource_type parameter is required")
			c.Abort()
			return
		}

		resource := model.ResourceType(resourceTypeStr)

		// Get resource ID if provided
		var resourceID *uint
		if idStr := c.Param("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				resourceIDUint := uint(id)
				resourceID = &resourceIDUint
			}
		}

		// Check permission
		permissionService := service.NewPermissionService()
		permissionCheck, err := permissionService.CheckPermission(userID.(uint), resource, action, resourceID)
		if err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check permission", err.Error())
			c.Abort()
			return
		}

		if !permissionCheck.HasPermission {
			response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", permissionCheck.Reason)
			c.Abort()
			return
		}

		c.Set("permission_source", permissionCheck.Source)
		c.Set("permission_resource", resource)
		c.Set("permission_action", action)

		c.Next()
	}
}

// AuditPermissionMiddleware logs permission checks for audit purposes
func AuditPermissionMiddleware(resource model.ResourceType, action model.PermissionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
			c.Abort()
			return
		}

		// Get request info for audit
		_ = c.ClientIP()
		_ = c.GetHeader("User-Agent")

		// Get resource ID if provided
		var resourceID *uint
		if idStr := c.Param("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				resourceIDUint := uint(id)
				resourceID = &resourceIDUint
			}
		}

		// Check permission
		permissionService := service.NewPermissionService()
		permissionCheck, err := permissionService.CheckPermission(userID.(uint), resource, action, resourceID)
		if err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check permission", err.Error())
			c.Abort()
			return
		}

		// Log the permission check
		details := map[string]interface{}{
			"resource":       resource,
			"action":         action,
			"has_permission": permissionCheck.HasPermission,
			"source":         permissionCheck.Source,
		}
		if resourceID != nil {
			details["resource_id"] = *resourceID
		}

		// Note: In a real implementation, you might want to log this asynchronously
		// to avoid impacting performance
		_, _, _ = permissionService.GetPermissionLogs(1, 1, map[string]interface{}{
			"user_id": userID,
		})

		if !permissionCheck.HasPermission {
			response.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", permissionCheck.Reason)
			c.Abort()
			return
		}

		c.Set("permission_source", permissionCheck.Source)
		c.Set("permission_resource", resource)
		c.Set("permission_action", action)

		c.Next()
	}
}

// Helper function to create permission requirements
func NewPermissionRequirement(resource model.ResourceType, action model.PermissionType, resourceIDParam string) PermissionRequirement {
	return PermissionRequirement{
		Resource:        resource,
		Action:          action,
		ResourceIDParam: resourceIDParam,
	}
}

// Common permission requirements
var (
	// User permissions
	UserReadPermission   = NewPermissionRequirement(model.ResourceTypeUser, model.PermissionTypeRead, "")
	UserWritePermission  = NewPermissionRequirement(model.ResourceTypeUser, model.PermissionTypeWrite, "")
	UserManagePermission = NewPermissionRequirement(model.ResourceTypeUser, model.PermissionTypeManage, "")
	UserAdminPermission  = NewPermissionRequirement(model.ResourceTypeUser, model.PermissionTypeAdmin, "")

	// Brand permissions
	BrandReadPermission   = NewPermissionRequirement(model.ResourceTypeBrand, model.PermissionTypeRead, "")
	BrandWritePermission  = NewPermissionRequirement(model.ResourceTypeBrand, model.PermissionTypeWrite, "")
	BrandManagePermission = NewPermissionRequirement(model.ResourceTypeBrand, model.PermissionTypeManage, "")
	BrandAdminPermission  = NewPermissionRequirement(model.ResourceTypeBrand, model.PermissionTypeAdmin, "")

	// Category permissions
	CategoryReadPermission   = NewPermissionRequirement(model.ResourceTypeCategory, model.PermissionTypeRead, "")
	CategoryWritePermission  = NewPermissionRequirement(model.ResourceTypeCategory, model.PermissionTypeWrite, "")
	CategoryManagePermission = NewPermissionRequirement(model.ResourceTypeCategory, model.PermissionTypeManage, "")
	CategoryAdminPermission  = NewPermissionRequirement(model.ResourceTypeCategory, model.PermissionTypeAdmin, "")

	// Product permissions
	ProductReadPermission   = NewPermissionRequirement(model.ResourceTypeProduct, model.PermissionTypeRead, "")
	ProductWritePermission  = NewPermissionRequirement(model.ResourceTypeProduct, model.PermissionTypeWrite, "")
	ProductManagePermission = NewPermissionRequirement(model.ResourceTypeProduct, model.PermissionTypeManage, "")
	ProductAdminPermission  = NewPermissionRequirement(model.ResourceTypeProduct, model.PermissionTypeAdmin, "")

	// Inventory permissions
	InventoryReadPermission   = NewPermissionRequirement(model.ResourceTypeInventory, model.PermissionTypeRead, "")
	InventoryWritePermission  = NewPermissionRequirement(model.ResourceTypeInventory, model.PermissionTypeWrite, "")
	InventoryManagePermission = NewPermissionRequirement(model.ResourceTypeInventory, model.PermissionTypeManage, "")
	InventoryAdminPermission  = NewPermissionRequirement(model.ResourceTypeInventory, model.PermissionTypeAdmin, "")

	// Upload permissions
	UploadReadPermission   = NewPermissionRequirement(model.ResourceTypeUpload, model.PermissionTypeRead, "")
	UploadWritePermission  = NewPermissionRequirement(model.ResourceTypeUpload, model.PermissionTypeWrite, "")
	UploadManagePermission = NewPermissionRequirement(model.ResourceTypeUpload, model.PermissionTypeManage, "")
	UploadAdminPermission  = NewPermissionRequirement(model.ResourceTypeUpload, model.PermissionTypeAdmin, "")

	// System permissions
	SystemReadPermission   = NewPermissionRequirement(model.ResourceTypeSystem, model.PermissionTypeRead, "")
	SystemWritePermission  = NewPermissionRequirement(model.ResourceTypeSystem, model.PermissionTypeWrite, "")
	SystemManagePermission = NewPermissionRequirement(model.ResourceTypeSystem, model.PermissionTypeManage, "")
	SystemAdminPermission  = NewPermissionRequirement(model.ResourceTypeSystem, model.PermissionTypeAdmin, "")
)
