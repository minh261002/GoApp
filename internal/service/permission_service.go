package service

import (
	"errors"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
)

// PermissionService defines methods for permission business logic
type PermissionService interface {
	// Permissions
	CreatePermission(req *model.PermissionCreateRequest, userID uint) (*model.PermissionResponse, error)
	GetPermissionByID(id uint) (*model.PermissionResponse, error)
	GetPermissionByName(name string) (*model.PermissionResponse, error)
	GetAllPermissions(page, limit int, filters map[string]interface{}) ([]model.PermissionResponse, int64, error)
	UpdatePermission(id uint, req *model.PermissionUpdateRequest, userID uint) (*model.PermissionResponse, error)
	DeletePermission(id uint, userID uint) error
	GetPermissionsByResource(resource model.ResourceType) ([]model.PermissionResponse, error)

	// Roles
	CreateRole(req *model.RoleCreateRequest, userID uint) (*model.RoleResponse, error)
	GetRoleByID(id uint) (*model.RoleResponse, error)
	GetRoleByName(name string) (*model.RoleResponse, error)
	GetAllRoles(page, limit int, filters map[string]interface{}) ([]model.RoleResponse, int64, error)
	UpdateRole(id uint, req *model.RoleUpdateRequest, userID uint) (*model.RoleResponse, error)
	DeleteRole(id uint, userID uint) error

	// Role Permissions
	AssignPermissionToRole(roleID, permissionID, userID uint) error
	RevokePermissionFromRole(roleID, permissionID, userID uint) error
	GetRolePermissions(roleID uint) ([]model.PermissionResponse, error)
	UpdateRolePermissions(roleID uint, req *model.RolePermissionRequest, userID uint) error

	// User Permissions
	AssignPermissionToUser(userID, permissionID, grantedBy uint, req *model.UserPermissionRequest) error
	RevokePermissionFromUser(userID, permissionID, grantedBy uint) error
	GetUserPermissions(userID uint) ([]model.UserPermissionResponse, error)
	GetUserEffectivePermissions(userID uint) ([]model.PermissionResponse, error)

	// Permission Checking
	CheckPermission(userID uint, resource model.ResourceType, action model.PermissionType, resourceID *uint) (*model.PermissionCheckResponse, error)
	GetUserPermissionsForResource(userID uint, resource model.ResourceType) ([]model.PermissionResponse, error)

	// Audit & Logging
	GetPermissionLogs(page, limit int, filters map[string]interface{}) ([]model.PermissionLog, int64, error)

	// Statistics
	GetPermissionStats() (*model.PermissionStatsResponse, error)
	GetUserPermissionStats(userID uint) (map[string]interface{}, error)

	// Utility
	InitializeDefaultPermissions() error
	SyncUserRole(userID uint, roleName string) error
}

// permissionService implements PermissionService
type permissionService struct {
	permissionRepo repository.PermissionRepository
	userRepo       repository.UserRepository
}

// NewPermissionService creates a new PermissionService
func NewPermissionService() PermissionService {
	return &permissionService{
		permissionRepo: repository.NewPermissionRepository(),
		userRepo:       repository.NewUserRepository(),
	}
}

// Permissions

// CreatePermission creates a new permission
func (s *permissionService) CreatePermission(req *model.PermissionCreateRequest, userID uint) (*model.PermissionResponse, error) {
	// Check if permission already exists
	existing, err := s.permissionRepo.GetPermissionByName(req.Name)
	if err != nil {
		logger.Errorf("Error checking existing permission: %v", err)
		return nil, fmt.Errorf("failed to check existing permission")
	}
	if existing != nil {
		return nil, errors.New("permission already exists")
	}

	permission := &model.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
		IsSystem:    false,
		IsActive:    true,
	}

	if err := s.permissionRepo.CreatePermission(permission); err != nil {
		logger.Errorf("Error creating permission: %v", err)
		return nil, fmt.Errorf("failed to create permission")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "create", "permission", permission.ID,
		fmt.Sprintf("Created permission: %s", permission.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return s.toPermissionResponse(permission), nil
}

// GetPermissionByID retrieves a permission by its ID
func (s *permissionService) GetPermissionByID(id uint) (*model.PermissionResponse, error) {
	permission, err := s.permissionRepo.GetPermissionByID(id)
	if err != nil {
		logger.Errorf("Error getting permission by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return nil, errors.New("permission not found")
	}
	return s.toPermissionResponse(permission), nil
}

// GetPermissionByName retrieves a permission by its name
func (s *permissionService) GetPermissionByName(name string) (*model.PermissionResponse, error) {
	permission, err := s.permissionRepo.GetPermissionByName(name)
	if err != nil {
		logger.Errorf("Error getting permission by name %s: %v", name, err)
		return nil, fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return nil, errors.New("permission not found")
	}
	return s.toPermissionResponse(permission), nil
}

// GetAllPermissions retrieves all permissions with pagination and filters
func (s *permissionService) GetAllPermissions(page, limit int, filters map[string]interface{}) ([]model.PermissionResponse, int64, error) {
	permissions, total, err := s.permissionRepo.GetAllPermissions(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting permissions: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve permissions")
	}

	var responses []model.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, *s.toPermissionResponse(&permission))
	}
	return responses, total, nil
}

// UpdatePermission updates an existing permission
func (s *permissionService) UpdatePermission(id uint, req *model.PermissionUpdateRequest, userID uint) (*model.PermissionResponse, error) {
	permission, err := s.permissionRepo.GetPermissionByID(id)
	if err != nil {
		logger.Errorf("Error getting permission by ID %d for update: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return nil, errors.New("permission not found")
	}

	// Check if permission is system permission
	if permission.IsSystem {
		return nil, errors.New("cannot update system permission")
	}

	// Update fields
	if req.DisplayName != "" {
		permission.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		permission.Description = req.Description
	}
	if req.IsActive != nil {
		permission.IsActive = *req.IsActive
	}

	if err := s.permissionRepo.UpdatePermission(permission); err != nil {
		logger.Errorf("Error updating permission %d: %v", id, err)
		return nil, fmt.Errorf("failed to update permission")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "update", "permission", permission.ID,
		fmt.Sprintf("Updated permission: %s", permission.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return s.toPermissionResponse(permission), nil
}

// DeletePermission deletes a permission
func (s *permissionService) DeletePermission(id uint, userID uint) error {
	permission, err := s.permissionRepo.GetPermissionByID(id)
	if err != nil {
		logger.Errorf("Error getting permission by ID %d for deletion: %v", id, err)
		return fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return errors.New("permission not found")
	}

	// Check if permission is system permission
	if permission.IsSystem {
		return errors.New("cannot delete system permission")
	}

	if err := s.permissionRepo.DeletePermission(id); err != nil {
		logger.Errorf("Error deleting permission %d: %v", id, err)
		return fmt.Errorf("failed to delete permission")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "delete", "permission", permission.ID,
		fmt.Sprintf("Deleted permission: %s", permission.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return nil
}

// GetPermissionsByResource retrieves permissions by resource
func (s *permissionService) GetPermissionsByResource(resource model.ResourceType) ([]model.PermissionResponse, error) {
	permissions, err := s.permissionRepo.GetPermissionsByResource(resource)
	if err != nil {
		logger.Errorf("Error getting permissions by resource %s: %v", resource, err)
		return nil, fmt.Errorf("failed to retrieve permissions")
	}

	var responses []model.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, *s.toPermissionResponse(&permission))
	}
	return responses, nil
}

// Roles

// CreateRole creates a new role
func (s *permissionService) CreateRole(req *model.RoleCreateRequest, userID uint) (*model.RoleResponse, error) {
	// Check if role already exists
	existing, err := s.permissionRepo.GetRoleByName(req.Name)
	if err != nil {
		logger.Errorf("Error checking existing role: %v", err)
		return nil, fmt.Errorf("failed to check existing role")
	}
	if existing != nil {
		return nil, errors.New("role already exists")
	}

	role := &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		IsActive:    true,
	}

	if err := s.permissionRepo.CreateRole(role); err != nil {
		logger.Errorf("Error creating role: %v", err)
		return nil, fmt.Errorf("failed to create role")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "create", "role", role.ID,
		fmt.Sprintf("Created role: %s", role.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return s.toRoleResponse(role), nil
}

// GetRoleByID retrieves a role by its ID
func (s *permissionService) GetRoleByID(id uint) (*model.RoleResponse, error) {
	role, err := s.permissionRepo.GetRoleByID(id)
	if err != nil {
		logger.Errorf("Error getting role by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve role")
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	return s.toRoleResponse(role), nil
}

// GetRoleByName retrieves a role by its name
func (s *permissionService) GetRoleByName(name string) (*model.RoleResponse, error) {
	role, err := s.permissionRepo.GetRoleByName(name)
	if err != nil {
		logger.Errorf("Error getting role by name %s: %v", name, err)
		return nil, fmt.Errorf("failed to retrieve role")
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	return s.toRoleResponse(role), nil
}

// GetAllRoles retrieves all roles with pagination and filters
func (s *permissionService) GetAllRoles(page, limit int, filters map[string]interface{}) ([]model.RoleResponse, int64, error) {
	roles, total, err := s.permissionRepo.GetAllRoles(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting roles: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve roles")
	}

	var responses []model.RoleResponse
	for _, role := range roles {
		responses = append(responses, *s.toRoleResponse(&role))
	}
	return responses, total, nil
}

// UpdateRole updates an existing role
func (s *permissionService) UpdateRole(id uint, req *model.RoleUpdateRequest, userID uint) (*model.RoleResponse, error) {
	role, err := s.permissionRepo.GetRoleByID(id)
	if err != nil {
		logger.Errorf("Error getting role by ID %d for update: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve role")
	}
	if role == nil {
		return nil, errors.New("role not found")
	}

	// Check if role is system role
	if role.IsSystem {
		return nil, errors.New("cannot update system role")
	}

	// Update fields
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	if err := s.permissionRepo.UpdateRole(role); err != nil {
		logger.Errorf("Error updating role %d: %v", id, err)
		return nil, fmt.Errorf("failed to update role")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "update", "role", role.ID,
		fmt.Sprintf("Updated role: %s", role.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return s.toRoleResponse(role), nil
}

// DeleteRole deletes a role
func (s *permissionService) DeleteRole(id uint, userID uint) error {
	role, err := s.permissionRepo.GetRoleByID(id)
	if err != nil {
		logger.Errorf("Error getting role by ID %d for deletion: %v", id, err)
		return fmt.Errorf("failed to retrieve role")
	}
	if role == nil {
		return errors.New("role not found")
	}

	// Check if role is system role
	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	if err := s.permissionRepo.DeleteRole(id); err != nil {
		logger.Errorf("Error deleting role %d: %v", id, err)
		return fmt.Errorf("failed to delete role")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "delete", "role", role.ID,
		fmt.Sprintf("Deleted role: %s", role.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return nil
}

// Role Permissions

// AssignPermissionToRole assigns a permission to a role
func (s *permissionService) AssignPermissionToRole(roleID, permissionID, userID uint) error {
	// Check if role exists
	role, err := s.permissionRepo.GetRoleByID(roleID)
	if err != nil {
		logger.Errorf("Error getting role by ID %d: %v", roleID, err)
		return fmt.Errorf("failed to retrieve role")
	}
	if role == nil {
		return errors.New("role not found")
	}

	// Check if permission exists
	permission, err := s.permissionRepo.GetPermissionByID(permissionID)
	if err != nil {
		logger.Errorf("Error getting permission by ID %d: %v", permissionID, err)
		return fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return errors.New("permission not found")
	}

	if err := s.permissionRepo.AssignPermissionToRole(roleID, permissionID, userID); err != nil {
		logger.Errorf("Error assigning permission to role: %v", err)
		return fmt.Errorf("failed to assign permission to role")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "grant", "role_permission", roleID,
		fmt.Sprintf("Assigned permission %s to role %s", permission.Name, role.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return nil
}

// RevokePermissionFromRole revokes a permission from a role
func (s *permissionService) RevokePermissionFromRole(roleID, permissionID, userID uint) error {
	// Check if role exists
	role, err := s.permissionRepo.GetRoleByID(roleID)
	if err != nil {
		logger.Errorf("Error getting role by ID %d: %v", roleID, err)
		return fmt.Errorf("failed to retrieve role")
	}
	if role == nil {
		return errors.New("role not found")
	}

	// Check if permission exists
	permission, err := s.permissionRepo.GetPermissionByID(permissionID)
	if err != nil {
		logger.Errorf("Error getting permission by ID %d: %v", permissionID, err)
		return fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return errors.New("permission not found")
	}

	if err := s.permissionRepo.RevokePermissionFromRole(roleID, permissionID); err != nil {
		logger.Errorf("Error revoking permission from role: %v", err)
		return fmt.Errorf("failed to revoke permission from role")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&userID, nil, "revoke", "role_permission", roleID,
		fmt.Sprintf("Revoked permission %s from role %s", permission.Name, role.Name), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return nil
}

// GetRolePermissions retrieves permissions for a role
func (s *permissionService) GetRolePermissions(roleID uint) ([]model.PermissionResponse, error) {
	permissions, err := s.permissionRepo.GetRolePermissions(roleID)
	if err != nil {
		logger.Errorf("Error getting role permissions: %v", err)
		return nil, fmt.Errorf("failed to retrieve role permissions")
	}

	var responses []model.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, *s.toPermissionResponse(&permission))
	}
	return responses, nil
}

// UpdateRolePermissions updates permissions for a role
func (s *permissionService) UpdateRolePermissions(roleID uint, req *model.RolePermissionRequest, userID uint) error {
	// Check if role exists
	role, err := s.permissionRepo.GetRoleByID(roleID)
	if err != nil {
		logger.Errorf("Error getting role by ID %d: %v", roleID, err)
		return fmt.Errorf("failed to retrieve role")
	}
	if role == nil {
		return errors.New("role not found")
	}

	// Get current permissions
	currentPermissionIDs, err := s.permissionRepo.GetRolePermissionIDs(roleID)
	if err != nil {
		logger.Errorf("Error getting current role permissions: %v", err)
		return fmt.Errorf("failed to retrieve current role permissions")
	}

	// Create maps for easier comparison
	currentMap := make(map[uint]bool)
	for _, id := range currentPermissionIDs {
		currentMap[id] = true
	}

	newMap := make(map[uint]bool)
	for _, id := range req.PermissionIDs {
		newMap[id] = true
	}

	// Add new permissions
	for _, permissionID := range req.PermissionIDs {
		if !currentMap[permissionID] {
			if err := s.AssignPermissionToRole(roleID, permissionID, userID); err != nil {
				return err
			}
		}
	}

	// Remove old permissions
	for _, permissionID := range currentPermissionIDs {
		if !newMap[permissionID] {
			if err := s.RevokePermissionFromRole(roleID, permissionID, userID); err != nil {
				return err
			}
		}
	}

	return nil
}

// User Permissions

// AssignPermissionToUser assigns a permission to a user
func (s *permissionService) AssignPermissionToUser(userID, permissionID, grantedBy uint, req *model.UserPermissionRequest) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if permission exists
	permission, err := s.permissionRepo.GetPermissionByID(permissionID)
	if err != nil {
		logger.Errorf("Error getting permission by ID %d: %v", permissionID, err)
		return fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return errors.New("permission not found")
	}

	if err := s.permissionRepo.AssignPermissionToUser(userID, permissionID, grantedBy, req.Reason, req.ExpiresAt); err != nil {
		logger.Errorf("Error assigning permission to user: %v", err)
		return fmt.Errorf("failed to assign permission to user")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&grantedBy, &userID, "grant", "user_permission", userID,
		fmt.Sprintf("Assigned permission %s to user %s", permission.Name, user.Username), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return nil
}

// RevokePermissionFromUser revokes a permission from a user
func (s *permissionService) RevokePermissionFromUser(userID, permissionID, grantedBy uint) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if permission exists
	permission, err := s.permissionRepo.GetPermissionByID(permissionID)
	if err != nil {
		logger.Errorf("Error getting permission by ID %d: %v", permissionID, err)
		return fmt.Errorf("failed to retrieve permission")
	}
	if permission == nil {
		return errors.New("permission not found")
	}

	if err := s.permissionRepo.RevokePermissionFromUser(userID, permissionID); err != nil {
		logger.Errorf("Error revoking permission from user: %v", err)
		return fmt.Errorf("failed to revoke permission from user")
	}

	// Log the action
	if err := s.permissionRepo.LogPermissionAction(
		&grantedBy, &userID, "revoke", "user_permission", userID,
		fmt.Sprintf("Revoked permission %s from user %s", permission.Name, user.Username), "", "",
	); err != nil {
		logger.Warnf("Failed to log permission action: %v", err)
	}

	return nil
}

// GetUserPermissions retrieves permissions for a user
func (s *permissionService) GetUserPermissions(userID uint) ([]model.UserPermissionResponse, error) {
	userPermissions, err := s.permissionRepo.GetUserPermissions(userID)
	if err != nil {
		logger.Errorf("Error getting user permissions: %v", err)
		return nil, fmt.Errorf("failed to retrieve user permissions")
	}

	var responses []model.UserPermissionResponse
	for _, up := range userPermissions {
		responses = append(responses, *s.toUserPermissionResponse(&up))
	}
	return responses, nil
}

// GetUserEffectivePermissions retrieves all effective permissions for a user
func (s *permissionService) GetUserEffectivePermissions(userID uint) ([]model.PermissionResponse, error) {
	permissions, err := s.permissionRepo.GetUserEffectivePermissions(userID)
	if err != nil {
		logger.Errorf("Error getting user effective permissions: %v", err)
		return nil, fmt.Errorf("failed to retrieve user effective permissions")
	}

	var responses []model.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, *s.toPermissionResponse(&permission))
	}
	return responses, nil
}

// Permission Checking

// CheckPermission checks if a user has permission for a specific resource and action
func (s *permissionService) CheckPermission(userID uint, resource model.ResourceType, action model.PermissionType, resourceID *uint) (*model.PermissionCheckResponse, error) {
	hasPermission, source, err := s.permissionRepo.CheckPermission(userID, resource, action, resourceID)
	if err != nil {
		logger.Errorf("Error checking permission: %v", err)
		return nil, fmt.Errorf("failed to check permission")
	}

	response := &model.PermissionCheckResponse{
		HasPermission: hasPermission,
		Source:        source,
	}

	if !hasPermission {
		response.Reason = "User does not have the required permission"
	}

	return response, nil
}

// GetUserPermissionsForResource retrieves user permissions for a specific resource
func (s *permissionService) GetUserPermissionsForResource(userID uint, resource model.ResourceType) ([]model.PermissionResponse, error) {
	permissions, err := s.permissionRepo.GetUserPermissionsForResource(userID, resource)
	if err != nil {
		logger.Errorf("Error getting user permissions for resource: %v", err)
		return nil, fmt.Errorf("failed to retrieve user permissions for resource")
	}

	var responses []model.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, *s.toPermissionResponse(&permission))
	}
	return responses, nil
}

// Audit & Logging

// GetPermissionLogs retrieves permission logs
func (s *permissionService) GetPermissionLogs(page, limit int, filters map[string]interface{}) ([]model.PermissionLog, int64, error) {
	logs, total, err := s.permissionRepo.GetPermissionLogs(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting permission logs: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve permission logs")
	}
	return logs, total, nil
}

// Statistics

// GetPermissionStats retrieves permission statistics
func (s *permissionService) GetPermissionStats() (*model.PermissionStatsResponse, error) {
	stats, err := s.permissionRepo.GetPermissionStats()
	if err != nil {
		logger.Errorf("Error getting permission stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve permission statistics")
	}
	return stats, nil
}

// GetUserPermissionStats retrieves permission statistics for a specific user
func (s *permissionService) GetUserPermissionStats(userID uint) (map[string]interface{}, error) {
	stats, err := s.permissionRepo.GetUserPermissionStats(userID)
	if err != nil {
		logger.Errorf("Error getting user permission stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve user permission statistics")
	}
	return stats, nil
}

// Utility

// InitializeDefaultPermissions initializes default permissions and roles
func (s *permissionService) InitializeDefaultPermissions() error {
	// This would typically be called during system initialization
	// The migration already creates default permissions and roles
	logger.Info("Default permissions and roles are initialized via migration")
	return nil
}

// SyncUserRole synchronizes user role with permission system
func (s *permissionService) SyncUserRole(userID uint, roleName string) error {
	// This method can be used to sync user roles when they change
	// For now, we'll just log the action
	logger.Infof("Syncing user %d role to %s", userID, roleName)
	return nil
}

// Helper methods for converting models to responses

func (s *permissionService) toPermissionResponse(permission *model.Permission) *model.PermissionResponse {
	return &model.PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		DisplayName: permission.DisplayName,
		Description: permission.Description,
		Resource:    permission.Resource,
		Action:      permission.Action,
		IsSystem:    permission.IsSystem,
		IsActive:    permission.IsActive,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

func (s *permissionService) toRoleResponse(role *model.Role) *model.RoleResponse {
	response := &model.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		IsActive:    role.IsActive,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	// Add permissions if available
	if len(role.RolePermissions) > 0 {
		var permissions []model.PermissionResponse
		for _, rp := range role.RolePermissions {
			if rp.Permission != nil {
				permissions = append(permissions, *s.toPermissionResponse(rp.Permission))
			}
		}
		response.Permissions = permissions
	}

	return response
}

func (s *permissionService) toUserPermissionResponse(up *model.UserPermission) *model.UserPermissionResponse {
	response := &model.UserPermissionResponse{
		ID:         up.ID,
		UserID:     up.UserID,
		Permission: *s.toPermissionResponse(up.Permission),
		IsGranted:  up.IsGranted,
		Reason:     up.Reason,
		ExpiresAt:  up.ExpiresAt,
		GrantedBy:  up.GrantedBy,
		CreatedAt:  up.CreatedAt,
		UpdatedAt:  up.UpdatedAt,
	}

	if up.User != nil {
		response.UserName = up.User.Username
	}
	if up.GrantedByUser != nil {
		response.GrantedByName = up.GrantedByUser.Username
	}

	return response
}
