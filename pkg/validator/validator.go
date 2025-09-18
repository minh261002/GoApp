package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// Đăng ký custom tag cho validation
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Đăng ký custom validation functions
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("username", validateUsername)

	return &CustomValidator{validator: v}
}

// ValidateStruct validate một struct
func (cv *CustomValidator) ValidateStruct(s interface{}) error {
	return cv.validator.Struct(s)
}

// ValidateVar validate một biến đơn lẻ
func (cv *CustomValidator) ValidateVar(field interface{}, tag string) error {
	return cv.validator.Var(field, tag)
}

// GetValidationErrors trả về danh sách lỗi validation
func (cv *CustomValidator) GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   err.Value(),
				Message: getErrorMessage(err),
			})
		}
	}

	return validationErrors
}

type ValidationError struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// Custom validation functions

// validatePassword kiểm tra mật khẩu mạnh
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Ít nhất 8 ký tự
	if len(password) < 8 {
		return false
	}

	// Có ít nhất 1 chữ hoa, 1 chữ thường, 1 số
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}

// validatePhone kiểm tra số điện thoại Việt Nam
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Loại bỏ khoảng trắng và dấu gạch ngang
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Kiểm tra độ dài và format
	if len(phone) < 10 || len(phone) > 11 {
		return false
	}

	// Bắt đầu bằng 0 hoặc +84
	if !strings.HasPrefix(phone, "0") && !strings.HasPrefix(phone, "+84") {
		return false
	}

	// Chỉ chứa số
	for _, char := range phone {
		if char < '0' || char > '9' {
			if char != '+' {
				return false
			}
		}
	}

	return true
}

// validateUsername kiểm tra tên người dùng
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Độ dài từ 3-20 ký tự
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	// Chỉ chứa chữ cái, số, dấu gạch dưới và dấu gạch ngang
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	// Bắt đầu bằng chữ cái hoặc số
	firstChar := username[0]
	if !((firstChar >= 'a' && firstChar <= 'z') ||
		(firstChar >= 'A' && firstChar <= 'Z') ||
		(firstChar >= '0' && firstChar <= '9')) {
		return false
	}

	return true
}

// getErrorMessage trả về thông báo lỗi thân thiện
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s là bắt buộc", err.Field())
	case "email":
		return fmt.Sprintf("%s phải là email hợp lệ", err.Field())
	case "min":
		return fmt.Sprintf("%s phải có ít nhất %s ký tự", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s không được vượt quá %s ký tự", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("%s phải có đúng %s ký tự", err.Field(), err.Param())
	case "password":
		return fmt.Sprintf("%s phải có ít nhất 8 ký tự, bao gồm chữ hoa, chữ thường và số", err.Field())
	case "phone":
		return fmt.Sprintf("%s phải là số điện thoại hợp lệ", err.Field())
	case "username":
		return fmt.Sprintf("%s phải có 3-20 ký tự, chỉ chứa chữ cái, số, _ và -", err.Field())
	default:
		return fmt.Sprintf("%s không hợp lệ", err.Field())
	}
}

// Global validator instance
var defaultValidator = NewCustomValidator()

// ValidateStruct validates a struct using the default validator
func ValidateStruct(s interface{}) error {
	return defaultValidator.ValidateStruct(s)
}
