package model

import (
	"time"

	"gorm.io/gorm"
)

// Session represents user session model for managing single device login
type Session struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	Token     string         `json:"token" gorm:"uniqueIndex;size:255;not null"`
	DeviceID  string         `json:"device_id" gorm:"size:100;not null"`
	UserAgent string         `json:"user_agent" gorm:"size:500"`
	IPAddress string         `json:"ip_address" gorm:"size:45"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

// TableName returns the table name for Session model
func (Session) TableName() string {
	return "sessions"
}

// IsExpired checks if session is expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if session is valid (active and not expired)
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}
