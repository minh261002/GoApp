package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"regexp"
	"strings"
	"time"
)

// GenerateRandomString tạo chuỗi ngẫu nhiên
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomNumber tạo số ngẫu nhiên trong khoảng [min, max]
func GenerateRandomNumber(min, max int) (int, error) {
	if min >= max {
		return min, nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}

	return int(n.Int64()) + min, nil
}

// IsValidEmail kiểm tra email hợp lệ
func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsValidPhone kiểm tra số điện thoại Việt Nam
func IsValidPhone(phone string) bool {
	// Loại bỏ khoảng trắng và dấu gạch ngang
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Kiểm tra độ dài
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

// FormatPhone chuẩn hóa số điện thoại
func FormatPhone(phone string) string {
	// Loại bỏ khoảng trắng và dấu gạch ngang
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Nếu bắt đầu bằng +84, thay thế bằng 0
	if strings.HasPrefix(phone, "+84") {
		phone = "0" + phone[3:]
	}

	return phone
}

// TruncateString cắt ngắn chuỗi với độ dài tối đa
func TruncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return str[:maxLength] + "..."
}

// Contains kiểm tra slice có chứa phần tử không
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates loại bỏ phần tử trùng lặp trong slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetCurrentTimestamp lấy timestamp hiện tại
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentTimeString lấy thời gian hiện tại dạng string
func GetCurrentTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// ParseTimeString parse string thành time
func ParseTimeString(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}

// FormatTime format time thành string
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// IsEmptyString kiểm tra chuỗi rỗng
func IsEmptyString(str string) bool {
	return strings.TrimSpace(str) == ""
}

// ToTitleCase chuyển đổi thành title case
func ToTitleCase(str string) string {
	if str == "" {
		return str
	}

	words := strings.Fields(strings.ToLower(str))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// GenerateSlug tạo slug từ chuỗi
func GenerateSlug(str string) string {
	// Chuyển thành chữ thường
	str = strings.ToLower(str)

	// Loại bỏ ký tự đặc biệt, chỉ giữ chữ cái, số và dấu gạch ngang
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	str = reg.ReplaceAllString(str, "")

	// Thay thế khoảng trắng bằng dấu gạch ngang
	reg = regexp.MustCompile(`\s+`)
	str = reg.ReplaceAllString(str, "-")

	// Loại bỏ dấu gạch ngang ở đầu và cuối
	str = strings.Trim(str, "-")

	return str
}
