package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey []byte
}

func NewJWTManager() *JWTManager {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "default-secret-key" // Fallback for development
	}
	return &JWTManager{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken tạo JWT token mới
func (j *JWTManager) GenerateToken(userID uint, username, email string) (string, error) {
	expireHours := 24
	if hours := os.Getenv("JWT_EXPIRE_HOURS"); hours != "" {
		if h, err := time.ParseDuration(hours + "h"); err == nil {
			expireHours = int(h.Hours())
		}
	}

	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go_app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken xác thực JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken làm mới token
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Tạo token mới với thông tin từ token cũ
	return j.GenerateToken(claims.UserID, claims.Username, claims.Email)
}
