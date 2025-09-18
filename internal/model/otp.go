package model

import (
	"time"

	"gorm.io/gorm"
)

// OTPType represents the type of OTP
type OTPType string

const (
	OTPTypePasswordReset OTPType = "password_reset"
	OTPTypeEmailVerify   OTPType = "email_verify"
)

// OTP represents OTP model for password reset and email verification
type OTP struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	Email     string         `json:"email" gorm:"size:100;not null;index"`
	Code      string         `json:"code" gorm:"size:10;not null"`
	Type      OTPType        `json:"type" gorm:"size:20;not null"`
	IsUsed    bool           `json:"is_used" gorm:"default:false"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

// TableName returns the table name for OTP model
func (OTP) TableName() string {
	return "otps"
}

// IsExpired checks if OTP is expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid checks if OTP is valid (not used and not expired)
func (o *OTP) IsValid() bool {
	return !o.IsUsed && !o.IsExpired()
}
