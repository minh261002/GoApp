package repository

import (
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

type OTPRepository interface {
	Create(otp *model.OTP) error
	GetByCode(code string) (*model.OTP, error)
	GetByEmail(email string) ([]*model.OTP, error)
	GetByUserID(userID uint) ([]*model.OTP, error)
	GetValidByCode(code string) (*model.OTP, error)
	GetValidByEmail(email string, otpType model.OTPType) (*model.OTP, error)
	Update(otp *model.OTP) error
	Delete(id uint) error
	DeleteExpired() error
	DeleteByUserID(userID uint) error
	MarkAsUsed(code string) error
}

type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository() OTPRepository {
	return &otpRepository{
		db: database.GetDB(),
	}
}

func (r *otpRepository) Create(otp *model.OTP) error {
	return r.db.Create(otp).Error
}

func (r *otpRepository) GetByCode(code string) (*model.OTP, error) {
	var otp model.OTP
	err := r.db.Preload("User").Where("code = ?", code).First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) GetByEmail(email string) ([]*model.OTP, error) {
	var otps []*model.OTP
	err := r.db.Where("email = ?", email).Find(&otps).Error
	if err != nil {
		return nil, err
	}
	return otps, nil
}

func (r *otpRepository) GetByUserID(userID uint) ([]*model.OTP, error) {
	var otps []*model.OTP
	err := r.db.Where("user_id = ?", userID).Find(&otps).Error
	if err != nil {
		return nil, err
	}
	return otps, nil
}

func (r *otpRepository) GetValidByCode(code string) (*model.OTP, error) {
	var otp model.OTP
	err := r.db.Preload("User").Where("code = ? AND is_used = ? AND expires_at > ?",
		code, false, time.Now()).First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) GetValidByEmail(email string, otpType model.OTPType) (*model.OTP, error) {
	var otp model.OTP
	err := r.db.Preload("User").Where("email = ? AND type = ? AND is_used = ? AND expires_at > ?",
		email, otpType, false, time.Now()).First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) Update(otp *model.OTP) error {
	return r.db.Save(otp).Error
}

func (r *otpRepository) Delete(id uint) error {
	return r.db.Delete(&model.OTP{}, id).Error
}

func (r *otpRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.OTP{}).Error
}

func (r *otpRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.OTP{}).Error
}

func (r *otpRepository) MarkAsUsed(code string) error {
	return r.db.Model(&model.OTP{}).Where("code = ?", code).Update("is_used", true).Error
}
