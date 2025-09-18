package repository

import (
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(session *model.Session) error
	GetByToken(token string) (*model.Session, error)
	GetByUserID(userID uint) ([]*model.Session, error)
	GetActiveByUserID(userID uint) ([]*model.Session, error)
	GetByDeviceID(userID uint, deviceID string) (*model.Session, error)
	Update(session *model.Session) error
	Delete(id uint) error
	DeleteByToken(token string) error
	DeleteByUserID(userID uint) error
	DeleteExpired() error
	DeactivateAllUserSessions(userID uint) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository() SessionRepository {
	return &sessionRepository{
		db: database.GetDB(),
	}
}

func (r *sessionRepository) Create(session *model.Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) GetByToken(token string) (*model.Session, error) {
	var session model.Session
	err := r.db.Preload("User").Where("token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByUserID(userID uint) ([]*model.Session, error) {
	var sessions []*model.Session
	err := r.db.Where("user_id = ?", userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepository) GetActiveByUserID(userID uint) ([]*model.Session, error) {
	var sessions []*model.Session
	err := r.db.Where("user_id = ? AND is_active = ? AND expires_at > ?",
		userID, true, time.Now()).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepository) GetByDeviceID(userID uint, deviceID string) (*model.Session, error) {
	var session model.Session
	err := r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) Update(session *model.Session) error {
	return r.db.Save(session).Error
}

func (r *sessionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Session{}, id).Error
}

func (r *sessionRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.Session{}).Error
}

func (r *sessionRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.Session{}).Error
}

func (r *sessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.Session{}).Error
}

func (r *sessionRepository) DeactivateAllUserSessions(userID uint) error {
	return r.db.Model(&model.Session{}).Where("user_id = ?", userID).Update("is_active", false).Error
}
