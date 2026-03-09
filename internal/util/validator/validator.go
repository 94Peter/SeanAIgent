package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// 簡單的台灣手機號碼格式校驗
	phoneRegex = regexp.MustCompile(`^09\d{8}$`)
)

// ValidateName 驗證姓名：去除空格、長度限制、禁止純符號
func ValidateName(name string, min, max int) (string, bool) {
	trimmed := strings.TrimSpace(name)
	count := utf8.RuneCountInString(trimmed)
	if count < min || count > max {
		return trimmed, false
	}
	// 禁止純數字或純特殊符號 (至少要有一個字母或中文字)
	hasAlpha := false
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= 0x4E00 && r <= 0x9FFF) {
			hasAlpha = true
			break
		}
	}
	return trimmed, hasAlpha
}

// ValidatePhone 驗證電話格式
func ValidatePhone(phone string) (string, bool) {
	trimmed := strings.TrimSpace(phone)
	// 移除非數字字元 (例如有人輸入 0912-345-678)
	cleaned := ""
	for _, r := range trimmed {
		if r >= '0' && r <= '9' {
			cleaned += string(r)
		}
	}
	if !phoneRegex.MatchString(cleaned) {
		return cleaned, false
	}
	return cleaned, true
}

// SanitizeInput 基本的輸入清理，防止惡意腳本 (雖然 Templ 會處理，但資料庫端也應保持乾淨)
func SanitizeInput(input string) string {
	s := strings.TrimSpace(input)
	// 移除常見的 HTML 標籤
	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
	return s
}
