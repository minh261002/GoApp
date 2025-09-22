package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents user model
type User struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Username        string         `json:"username" gorm:"uniqueIndex;size:50;not null" validate:"required,username"`
	Email           string         `json:"email" gorm:"uniqueIndex;size:100;not null" validate:"required,email"`
	Password        string         `json:"-" gorm:"size:255;not null" validate:"required,password"`
	FirstName       string         `json:"first_name" gorm:"size:50" validate:"max=50"`
	LastName        string         `json:"last_name" gorm:"size:50" validate:"max=50"`
	Phone           string         `json:"phone" gorm:"size:20" validate:"phone"`
	Avatar          string         `json:"avatar" gorm:"size:255"`
	RoleID          uint           `json:"role_id" gorm:"not null;index"`
	UserRole        *Role          `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	IsEmailVerified bool           `json:"is_email_verified" gorm:"default:false"`
	LastLogin       *time.Time     `json:"last_login,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Sessions []Session `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	OTPs     []OTP     `json:"otps,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Set default role if not specified - get 'user' role from database
	if u.RoleID == 0 {
		var role Role
		if err := tx.Where("name = ? AND is_active = ?", "user", true).First(&role).Error; err == nil {
			u.RoleID = role.ID
		}
	}
	return nil
}

// GetFullName returns the full name of the user
func (u *User) GetFullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.Username
}

// IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.UserRole != nil && u.UserRole.Name == "admin"
}

// IsModerator checks if user is moderator or admin
func (u *User) IsModerator() bool {
	return u.UserRole != nil && (u.UserRole.Name == "moderator" || u.UserRole.Name == "admin")
}

// HasRole checks if user has a specific role
func (u *User) HasRole(roleName string) bool {
	return u.UserRole != nil && u.UserRole.Name == roleName
}

// HasAnyRole checks if user has any of the specified roles
func (u *User) HasAnyRole(roleNames ...string) bool {
	if u.UserRole == nil {
		return false
	}
	for _, roleName := range roleNames {
		if u.UserRole.Name == roleName {
			return true
		}
	}
	return false
}
