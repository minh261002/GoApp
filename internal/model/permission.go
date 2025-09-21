package model

import (
	"time"

	"gorm.io/gorm"
)

// PermissionType defines the type of permission
type PermissionType string

const (
	PermissionTypeRead   PermissionType = "read"   // Đọc dữ liệu
	PermissionTypeWrite  PermissionType = "write"  // Ghi dữ liệu
	PermissionTypeDelete PermissionType = "delete" // Xóa dữ liệu
	PermissionTypeManage PermissionType = "manage" // Quản lý
	PermissionTypeAdmin  PermissionType = "admin"  // Quản trị
)

// ResourceType defines the type of resource
type ResourceType string

const (
	ResourceTypeUser         ResourceType = "user"
	ResourceTypeBrand        ResourceType = "brand"
	ResourceTypeCategory     ResourceType = "category"
	ResourceTypeProduct      ResourceType = "product"
	ResourceTypeInventory    ResourceType = "inventory"
	ResourceTypeUpload       ResourceType = "upload"
	ResourceTypeOrder        ResourceType = "order"
	ResourceTypeAddress      ResourceType = "address"
	ResourceTypeReview       ResourceType = "review"
	ResourceTypeCoupon       ResourceType = "coupon"
	ResourceTypePoint        ResourceType = "point"
	ResourceTypeBanner       ResourceType = "banner"
	ResourceTypeSlider       ResourceType = "slider"
	ResourceTypeWishlist     ResourceType = "wishlist"
	ResourceTypeSearch       ResourceType = "search"
	ResourceTypeNotification ResourceType = "notification"
	ResourceTypeCustomer     ResourceType = "customer"
	ResourceTypeReport       ResourceType = "report"
	ResourceTypeSystem       ResourceType = "system"
	ResourceTypeAudit        ResourceType = "audit"
)

// Permission represents a permission in the system
type Permission struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:100;not null"` // e.g., "brands.read"
	DisplayName string         `json:"display_name" gorm:"size:255;not null"`     // e.g., "Read Brands"
	Description string         `json:"description" gorm:"type:text"`
	Resource    ResourceType   `json:"resource" gorm:"size:50;not null"` // e.g., "brand"
	Action      PermissionType `json:"action" gorm:"size:50;not null"`   // e.g., "read"
	IsSystem    bool           `json:"is_system" gorm:"default:false"`   // System permission (cannot be deleted)
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	RolePermissions []RolePermission `json:"role_permissions,omitempty" gorm:"foreignKey:PermissionID"`
	UserPermissions []UserPermission `json:"user_permissions,omitempty" gorm:"foreignKey:PermissionID"`
}

// Role represents a role in the system
type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:50;not null"` // e.g., "admin", "moderator", "user"
	DisplayName string         `json:"display_name" gorm:"size:100;not null"`    // e.g., "Administrator"
	Description string         `json:"description" gorm:"type:text"`
	IsSystem    bool           `json:"is_system" gorm:"default:false"` // System role (cannot be deleted)
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	RolePermissions []RolePermission `json:"role_permissions,omitempty" gorm:"foreignKey:RoleID"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	RoleID        uint           `json:"role_id" gorm:"not null;index"`
	Role          *Role          `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	PermissionID  uint           `json:"permission_id" gorm:"not null;index"`
	Permission    *Permission    `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
	GrantedBy     uint           `json:"granted_by" gorm:"not null;index"`
	GrantedByUser *User          `json:"granted_by_user,omitempty" gorm:"foreignKey:GrantedBy"`
	GrantedAt     time.Time      `json:"granted_at" gorm:"autoCreateTime"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_role_permission,role_id,permission_id"`
}

// UserPermission represents direct permissions assigned to users (overrides role permissions)
type UserPermission struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	User          *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	PermissionID  uint           `json:"permission_id" gorm:"not null;index"`
	Permission    *Permission    `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
	IsGranted     bool           `json:"is_granted" gorm:"default:true"` // true = grant, false = deny
	GrantedBy     uint           `json:"granted_by" gorm:"not null;index"`
	GrantedByUser *User          `json:"granted_by_user,omitempty" gorm:"foreignKey:GrantedBy"`
	Reason        string         `json:"reason" gorm:"type:text"` // Reason for granting/denying
	ExpiresAt     *time.Time     `json:"expires_at"`              // Optional expiration
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_user_permission,user_id,permission_id"`
}

// PermissionLog represents audit log for permission changes
type PermissionLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	User         *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action       string    `json:"action" gorm:"size:50;not null"`        // grant, revoke, create, update, delete
	ResourceType string    `json:"resource_type" gorm:"size:50;not null"` // role, permission, user_permission
	ResourceID   uint      `json:"resource_id" gorm:"not null"`
	Details      string    `json:"details" gorm:"type:text"` // JSON details
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	UserAgent    string    `json:"user_agent" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	TargetUserID *uint `json:"target_user_id" gorm:"index"`
	TargetUser   *User `json:"target_user,omitempty" gorm:"foreignKey:TargetUserID"`
}

// Request/Response structs

// PermissionCreateRequest represents the request body for creating a permission
type PermissionCreateRequest struct {
	Name        string         `json:"name" binding:"required,min=3,max=100"`
	DisplayName string         `json:"display_name" binding:"required,min=3,max=255"`
	Description string         `json:"description"`
	Resource    ResourceType   `json:"resource" binding:"required,oneof=user brand category product inventory upload order customer report system audit"`
	Action      PermissionType `json:"action" binding:"required,oneof=read write delete manage admin"`
}

// PermissionUpdateRequest represents the request body for updating a permission
type PermissionUpdateRequest struct {
	DisplayName string `json:"display_name" binding:"min=3,max=255"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

// RoleCreateRequest represents the request body for creating a role
type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=3,max=100"`
	Description string `json:"description"`
}

// RoleUpdateRequest represents the request body for updating a role
type RoleUpdateRequest struct {
	DisplayName string `json:"display_name" binding:"min=3,max=100"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

// RolePermissionRequest represents the request body for assigning permissions to role
type RolePermissionRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required,min=1"`
}

// UserPermissionRequest represents the request body for assigning permissions to user
type UserPermissionRequest struct {
	PermissionID uint       `json:"permission_id" binding:"required"`
	IsGranted    bool       `json:"is_granted"`
	Reason       string     `json:"reason"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

// PermissionResponse represents the response body for a permission
type PermissionResponse struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	Resource    ResourceType   `json:"resource"`
	Action      PermissionType `json:"action"`
	IsSystem    bool           `json:"is_system"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// RoleResponse represents the response body for a role
type RoleResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	DisplayName string               `json:"display_name"`
	Description string               `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	IsActive    bool                 `json:"is_active"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	UserCount   int64                `json:"user_count,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// UserPermissionResponse represents the response body for user permission
type UserPermissionResponse struct {
	ID            uint               `json:"id"`
	UserID        uint               `json:"user_id"`
	UserName      string             `json:"user_name,omitempty"`
	Permission    PermissionResponse `json:"permission"`
	IsGranted     bool               `json:"is_granted"`
	Reason        string             `json:"reason"`
	ExpiresAt     *time.Time         `json:"expires_at"`
	GrantedBy     uint               `json:"granted_by"`
	GrantedByName string             `json:"granted_by_name,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// PermissionCheckRequest represents the request body for checking permissions
type PermissionCheckRequest struct {
	UserID     uint           `json:"user_id" binding:"required"`
	Resource   ResourceType   `json:"resource" binding:"required"`
	Action     PermissionType `json:"action" binding:"required"`
	ResourceID *uint          `json:"resource_id,omitempty"` // For resource-specific permissions
}

// PermissionCheckResponse represents the response body for permission check
type PermissionCheckResponse struct {
	HasPermission bool   `json:"has_permission"`
	Source        string `json:"source"` // role, user_permission, system
	Reason        string `json:"reason,omitempty"`
}

// PermissionStatsResponse represents permission statistics
type PermissionStatsResponse struct {
	TotalPermissions      int64 `json:"total_permissions"`
	ActivePermissions     int64 `json:"active_permissions"`
	TotalRoles            int64 `json:"total_roles"`
	ActiveRoles           int64 `json:"active_roles"`
	TotalUserPermissions  int64 `json:"total_user_permissions"`
	ActiveUserPermissions int64 `json:"active_user_permissions"`
}

// Helper methods

// GetPermissionName generates permission name from resource and action
func GetPermissionName(resource ResourceType, action PermissionType) string {
	return string(resource) + "." + string(action)
}

// IsExpired checks if user permission is expired
func (up *UserPermission) IsExpired() bool {
	if up.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*up.ExpiresAt)
}

// IsValid checks if user permission is valid (not expired and active)
func (up *UserPermission) IsValid() bool {
	return !up.IsExpired() && up.IsGranted
}

// GetFullName returns the full permission name
func (p *Permission) GetFullName() string {
	return p.Name
}

// GetDisplayName returns the display name or name if display name is empty
func (p *Permission) GetDisplayName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Name
}

// GetFullName returns the full role name
func (r *Role) GetFullName() string {
	return r.Name
}

// GetDisplayName returns the display name or name if display name is empty
func (r *Role) GetDisplayName() string {
	if r.DisplayName != "" {
		return r.DisplayName
	}
	return r.Name
}
