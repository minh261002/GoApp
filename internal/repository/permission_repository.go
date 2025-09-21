package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// PermissionRepository defines methods for interacting with permission data
type PermissionRepository interface {
	// Permissions
	CreatePermission(permission *model.Permission) error
	GetPermissionByID(id uint) (*model.Permission, error)
	GetPermissionByName(name string) (*model.Permission, error)
	GetPermissionByResourceAction(resource model.ResourceType, action model.PermissionType) (*model.Permission, error)
	GetAllPermissions(page, limit int, filters map[string]interface{}) ([]model.Permission, int64, error)
	UpdatePermission(permission *model.Permission) error
	DeletePermission(id uint) error
	GetPermissionsByResource(resource model.ResourceType) ([]model.Permission, error)

	// Roles
	CreateRole(role *model.Role) error
	GetRoleByID(id uint) (*model.Role, error)
	GetRoleByName(name string) (*model.Role, error)
	GetAllRoles(page, limit int, filters map[string]interface{}) ([]model.Role, int64, error)
	UpdateRole(role *model.Role) error
	DeleteRole(id uint) error
	GetRolesByUser(userID uint) ([]model.Role, error)

	// Role Permissions
	AssignPermissionToRole(roleID, permissionID, grantedBy uint) error
	RevokePermissionFromRole(roleID, permissionID uint) error
	GetRolePermissions(roleID uint) ([]model.Permission, error)
	GetRolePermissionIDs(roleID uint) ([]uint, error)
	CheckRoleHasPermission(roleID uint, permissionName string) (bool, error)

	// User Permissions
	AssignPermissionToUser(userID, permissionID, grantedBy uint, reason string, expiresAt *time.Time) error
	RevokePermissionFromUser(userID, permissionID uint) error
	GetUserPermissions(userID uint) ([]model.UserPermission, error)
	GetUserPermissionIDs(userID uint) ([]uint, error)
	CheckUserHasPermission(userID uint, permissionName string) (bool, error)
	GetUserEffectivePermissions(userID uint) ([]model.Permission, error)

	// Permission Checking
	CheckPermission(userID uint, resource model.ResourceType, action model.PermissionType, resourceID *uint) (bool, string, error)
	GetUserPermissionsForResource(userID uint, resource model.ResourceType) ([]model.Permission, error)

	// Audit Logging
	LogPermissionAction(userID, targetUserID *uint, action, resourceType string, resourceID uint, details, ipAddress, userAgent string) error
	GetPermissionLogs(page, limit int, filters map[string]interface{}) ([]model.PermissionLog, int64, error)

	// Statistics
	GetPermissionStats() (*model.PermissionStatsResponse, error)
	GetUserPermissionStats(userID uint) (map[string]interface{}, error)
}

// permissionRepository implements PermissionRepository
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new PermissionRepository
func NewPermissionRepository() PermissionRepository {
	return &permissionRepository{
		db: database.DB,
	}
}

// Permissions

// CreatePermission creates a new permission
func (r *permissionRepository) CreatePermission(permission *model.Permission) error {
	return r.db.Create(permission).Error
}

// GetPermissionByID retrieves a permission by its ID
func (r *permissionRepository) GetPermissionByID(id uint) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.First(&permission, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByName retrieves a permission by its name
func (r *permissionRepository) GetPermissionByName(name string) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.Where("name = ?", name).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByResourceAction retrieves a permission by resource and action
func (r *permissionRepository) GetPermissionByResourceAction(resource model.ResourceType, action model.PermissionType) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.Where("resource = ? AND action = ?", resource, action).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// GetAllPermissions retrieves all permissions with pagination and filters
func (r *permissionRepository) GetAllPermissions(page, limit int, filters map[string]interface{}) ([]model.Permission, int64, error) {
	var permissions []model.Permission
	var total int64
	db := r.db.Model(&model.Permission{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "resource":
			db = db.Where("resource = ?", value)
		case "action":
			db = db.Where("action = ?", value)
		case "is_active":
			db = db.Where("is_active = ?", value)
		case "is_system":
			db = db.Where("is_system = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("name LIKE ? OR display_name LIKE ? OR description LIKE ?", searchTerm, searchTerm, searchTerm)
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("resource, action")

	if err := db.Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// UpdatePermission updates an existing permission
func (r *permissionRepository) UpdatePermission(permission *model.Permission) error {
	return r.db.Save(permission).Error
}

// DeletePermission soft deletes a permission
func (r *permissionRepository) DeletePermission(id uint) error {
	// Check if permission is system permission
	var permission model.Permission
	if err := r.db.First(&permission, id).Error; err != nil {
		return err
	}
	if permission.IsSystem {
		return fmt.Errorf("cannot delete system permission")
	}
	return r.db.Delete(&model.Permission{}, id).Error
}

// GetPermissionsByResource retrieves permissions by resource
func (r *permissionRepository) GetPermissionsByResource(resource model.ResourceType) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Where("resource = ? AND is_active = ?", resource, true).Order("action").Find(&permissions).Error
	return permissions, err
}

// Roles

// CreateRole creates a new role
func (r *permissionRepository) CreateRole(role *model.Role) error {
	return r.db.Create(role).Error
}

// GetRoleByID retrieves a role by its ID
func (r *permissionRepository) GetRoleByID(id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.Preload("RolePermissions.Permission").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// GetRoleByName retrieves a role by its name
func (r *permissionRepository) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("name = ?", name).Preload("RolePermissions.Permission").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// GetAllRoles retrieves all roles with pagination and filters
func (r *permissionRepository) GetAllRoles(page, limit int, filters map[string]interface{}) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	db := r.db.Model(&model.Role{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "is_active":
			db = db.Where("is_active = ?", value)
		case "is_system":
			db = db.Where("is_system = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("name LIKE ? OR display_name LIKE ? OR description LIKE ?", searchTerm, searchTerm, searchTerm)
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("name")

	if err := db.Preload("RolePermissions.Permission").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// UpdateRole updates an existing role
func (r *permissionRepository) UpdateRole(role *model.Role) error {
	return r.db.Save(role).Error
}

// DeleteRole soft deletes a role
func (r *permissionRepository) DeleteRole(id uint) error {
	// Check if role is system role
	var role model.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return err
	}
	if role.IsSystem {
		return fmt.Errorf("cannot delete system role")
	}
	return r.db.Delete(&model.Role{}, id).Error
}

// GetRolesByUser retrieves roles for a user
func (r *permissionRepository) GetRolesByUser(userID uint) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Table("roles").
		Joins("JOIN users ON roles.name = users.role").
		Where("users.id = ? AND roles.is_active = ?", userID, true).
		Find(&roles).Error
	return roles, err
}

// Role Permissions

// AssignPermissionToRole assigns a permission to a role
func (r *permissionRepository) AssignPermissionToRole(roleID, permissionID, grantedBy uint) error {
	// Check if assignment already exists
	var existing model.RolePermission
	err := r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).First(&existing).Error
	if err == nil {
		return fmt.Errorf("permission already assigned to role")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	rolePermission := &model.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		GrantedBy:    grantedBy,
		GrantedAt:    time.Now(),
	}

	return r.db.Create(rolePermission).Error
}

// RevokePermissionFromRole revokes a permission from a role
func (r *permissionRepository) RevokePermissionFromRole(roleID, permissionID uint) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&model.RolePermission{}).Error
}

// GetRolePermissions retrieves permissions for a role
func (r *permissionRepository) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.is_active = ?", roleID, true).
		Order("permissions.resource, permissions.action").
		Find(&permissions).Error
	return permissions, err
}

// GetRolePermissionIDs retrieves permission IDs for a role
func (r *permissionRepository) GetRolePermissionIDs(roleID uint) ([]uint, error) {
	var permissionIDs []uint
	err := r.db.Table("role_permissions").
		Select("permission_id").
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permissionIDs).Error
	return permissionIDs, err
}

// CheckRoleHasPermission checks if a role has a specific permission
func (r *permissionRepository) CheckRoleHasPermission(roleID uint, permissionName string) (bool, error) {
	var count int64
	err := r.db.Table("role_permissions").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ? AND permissions.name = ? AND permissions.is_active = ?", roleID, permissionName, true).
		Count(&count).Error
	return count > 0, err
}

// User Permissions

// AssignPermissionToUser assigns a permission to a user
func (r *permissionRepository) AssignPermissionToUser(userID, permissionID, grantedBy uint, reason string, expiresAt *time.Time) error {
	// Check if assignment already exists
	var existing model.UserPermission
	err := r.db.Where("user_id = ? AND permission_id = ?", userID, permissionID).First(&existing).Error
	if err == nil {
		// Update existing assignment
		existing.IsGranted = true
		existing.Reason = reason
		existing.ExpiresAt = expiresAt
		existing.GrantedBy = grantedBy
		return r.db.Save(&existing).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	userPermission := &model.UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
		IsGranted:    true,
		Reason:       reason,
		ExpiresAt:    expiresAt,
		GrantedBy:    grantedBy,
	}

	return r.db.Create(userPermission).Error
}

// RevokePermissionFromUser revokes a permission from a user
func (r *permissionRepository) RevokePermissionFromUser(userID, permissionID uint) error {
	return r.db.Where("user_id = ? AND permission_id = ?", userID, permissionID).Delete(&model.UserPermission{}).Error
}

// GetUserPermissions retrieves permissions for a user
func (r *permissionRepository) GetUserPermissions(userID uint) ([]model.UserPermission, error) {
	var userPermissions []model.UserPermission
	err := r.db.Where("user_id = ?", userID).
		Preload("Permission").
		Preload("GrantedByUser").
		Order("created_at DESC").
		Find(&userPermissions).Error
	return userPermissions, err
}

// GetUserPermissionIDs retrieves permission IDs for a user
func (r *permissionRepository) GetUserPermissionIDs(userID uint) ([]uint, error) {
	var permissionIDs []uint
	err := r.db.Table("user_permissions").
		Select("permission_id").
		Where("user_id = ? AND is_granted = ? AND (expires_at IS NULL OR expires_at > ?)", userID, true, time.Now()).
		Pluck("permission_id", &permissionIDs).Error
	return permissionIDs, err
}

// CheckUserHasPermission checks if a user has a specific permission
func (r *permissionRepository) CheckUserHasPermission(userID uint, permissionName string) (bool, error) {
	// Check user permissions first (direct assignments)
	var userPermissionCount int64
	err := r.db.Table("user_permissions").
		Joins("JOIN permissions ON user_permissions.permission_id = permissions.id").
		Where("user_permissions.user_id = ? AND permissions.name = ? AND user_permissions.is_granted = ? AND permissions.is_active = ? AND (user_permissions.expires_at IS NULL OR user_permissions.expires_at > ?)",
			userID, permissionName, true, true, time.Now()).
		Count(&userPermissionCount).Error
	if err != nil {
		return false, err
	}
	if userPermissionCount > 0 {
		return true, nil
	}

	// Check role permissions
	var rolePermissionCount int64
	err = r.db.Table("role_permissions").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN users ON users.role = roles.name").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Where("users.id = ? AND permissions.name = ? AND permissions.is_active = ? AND roles.is_active = ?",
			userID, permissionName, true, true).
		Count(&rolePermissionCount).Error
	if err != nil {
		return false, err
	}

	return rolePermissionCount > 0, nil
}

// GetUserEffectivePermissions retrieves all effective permissions for a user
func (r *permissionRepository) GetUserEffectivePermissions(userID uint) ([]model.Permission, error) {
	var permissions []model.Permission

	// Get user's role
	var user model.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Get permissions from role
	rolePermissions, err := r.GetRolePermissionsByRoleName(user.UserRole.Name)
	if err != nil {
		return nil, err
	}
	permissions = append(permissions, rolePermissions...)

	// Get direct user permissions
	var userPermissions []model.Permission
	err = r.db.Table("permissions").
		Joins("JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ? AND user_permissions.is_granted = ? AND permissions.is_active = ? AND (user_permissions.expires_at IS NULL OR user_permissions.expires_at > ?)",
			userID, true, true, time.Now()).
		Find(&userPermissions).Error
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	permissionMap := make(map[uint]model.Permission)
	for _, p := range permissions {
		permissionMap[p.ID] = p
	}
	for _, p := range userPermissions {
		permissionMap[p.ID] = p
	}

	var result []model.Permission
	for _, p := range permissionMap {
		result = append(result, p)
	}

	return result, nil
}

// GetRolePermissionsByRoleName retrieves permissions by role name
func (r *permissionRepository) GetRolePermissionsByRoleName(roleName string) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Where("roles.name = ? AND permissions.is_active = ? AND roles.is_active = ?", roleName, true, true).
		Order("permissions.resource, permissions.action").
		Find(&permissions).Error
	return permissions, err
}

// Permission Checking

// CheckPermission checks if a user has permission for a specific resource and action
func (r *permissionRepository) CheckPermission(userID uint, resource model.ResourceType, action model.PermissionType, resourceID *uint) (bool, string, error) {
	permissionName := model.GetPermissionName(resource, action)

	// Check user permissions first (direct assignments)
	var userPermission model.UserPermission
	err := r.db.Table("user_permissions").
		Joins("JOIN permissions ON user_permissions.permission_id = permissions.id").
		Where("user_permissions.user_id = ? AND permissions.name = ? AND user_permissions.is_granted = ? AND permissions.is_active = ? AND (user_permissions.expires_at IS NULL OR user_permissions.expires_at > ?)",
			userID, permissionName, true, true, time.Now()).
		First(&userPermission).Error
	if err == nil {
		return true, "user_permission", nil
	}
	if err != gorm.ErrRecordNotFound {
		return false, "", err
	}

	// Check role permissions
	var rolePermission model.RolePermission
	err = r.db.Table("role_permissions").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN users ON users.role = roles.name").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Where("users.id = ? AND permissions.name = ? AND permissions.is_active = ? AND roles.is_active = ?",
			userID, permissionName, true, true).
		First(&rolePermission).Error
	if err == nil {
		return true, "role", nil
	}
	if err != gorm.ErrRecordNotFound {
		return false, "", err
	}

	return false, "", nil
}

// GetUserPermissionsForResource retrieves user permissions for a specific resource
func (r *permissionRepository) GetUserPermissionsForResource(userID uint, resource model.ResourceType) ([]model.Permission, error) {
	var permissions []model.Permission

	// Get permissions from role
	var user model.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	rolePermissions, err := r.GetRolePermissionsByRoleName(user.UserRole.Name)
	if err != nil {
		return nil, err
	}

	for _, p := range rolePermissions {
		if p.Resource == resource {
			permissions = append(permissions, p)
		}
	}

	// Get direct user permissions for resource
	var userPermissions []model.Permission
	err = r.db.Table("permissions").
		Joins("JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ? AND permissions.resource = ? AND user_permissions.is_granted = ? AND permissions.is_active = ? AND (user_permissions.expires_at IS NULL OR user_permissions.expires_at > ?)",
			userID, resource, true, true, time.Now()).
		Find(&userPermissions).Error
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	permissionMap := make(map[uint]model.Permission)
	for _, p := range permissions {
		permissionMap[p.ID] = p
	}
	for _, p := range userPermissions {
		permissionMap[p.ID] = p
	}

	var result []model.Permission
	for _, p := range permissionMap {
		result = append(result, p)
	}

	return result, nil
}

// Audit Logging

// LogPermissionAction logs a permission-related action
func (r *permissionRepository) LogPermissionAction(userID, targetUserID *uint, action, resourceType string, resourceID uint, details, ipAddress, userAgent string) error {
	log := &model.PermissionLog{
		UserID:       *userID,
		TargetUserID: targetUserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}

	return r.db.Create(log).Error
}

// GetPermissionLogs retrieves permission logs with pagination and filters
func (r *permissionRepository) GetPermissionLogs(page, limit int, filters map[string]interface{}) ([]model.PermissionLog, int64, error) {
	var logs []model.PermissionLog
	var total int64
	db := r.db.Model(&model.PermissionLog{}).Preload("User").Preload("TargetUser")

	// Apply filters
	for key, value := range filters {
		switch key {
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "target_user_id":
			db = db.Where("target_user_id = ?", value)
		case "action":
			db = db.Where("action = ?", value)
		case "resource_type":
			db = db.Where("resource_type = ?", value)
		case "date_from":
			db = db.Where("created_at >= ?", value)
		case "date_to":
			db = db.Where("created_at <= ?", value)
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("created_at DESC")

	if err := db.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Statistics

// GetPermissionStats retrieves permission statistics
func (r *permissionRepository) GetPermissionStats() (*model.PermissionStatsResponse, error) {
	var stats model.PermissionStatsResponse
	var count int64

	// Total permissions
	r.db.Model(&model.Permission{}).Count(&count)
	stats.TotalPermissions = count

	// Active permissions
	r.db.Model(&model.Permission{}).Where("is_active = ?", true).Count(&count)
	stats.ActivePermissions = count

	// Total roles
	r.db.Model(&model.Role{}).Count(&count)
	stats.TotalRoles = count

	// Active roles
	r.db.Model(&model.Role{}).Where("is_active = ?", true).Count(&count)
	stats.ActiveRoles = count

	// Total user permissions
	r.db.Model(&model.UserPermission{}).Count(&count)
	stats.TotalUserPermissions = count

	// Active user permissions
	r.db.Model(&model.UserPermission{}).Where("is_granted = ? AND (expires_at IS NULL OR expires_at > ?)", true, time.Now()).Count(&count)
	stats.ActiveUserPermissions = count

	return &stats, nil
}

// GetUserPermissionStats retrieves permission statistics for a specific user
func (r *permissionRepository) GetUserPermissionStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get user's role
	var user model.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	stats["role"] = user.UserRole.Name

	// Count role permissions
	rolePermissionCount, err := r.GetRolePermissionIDs(userID)
	if err != nil {
		return nil, err
	}
	stats["role_permissions_count"] = len(rolePermissionCount)

	// Count direct user permissions
	userPermissionCount, err := r.GetUserPermissionIDs(userID)
	if err != nil {
		return nil, err
	}
	stats["user_permissions_count"] = len(userPermissionCount)

	// Get effective permissions
	effectivePermissions, err := r.GetUserEffectivePermissions(userID)
	if err != nil {
		return nil, err
	}
	stats["effective_permissions_count"] = len(effectivePermissions)

	// Group permissions by resource
	resourceCounts := make(map[string]int)
	for _, p := range effectivePermissions {
		resourceCounts[string(p.Resource)]++
	}
	stats["permissions_by_resource"] = resourceCounts

	return stats, nil
}
