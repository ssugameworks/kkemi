package utils

import (
	"discord-bot/constants"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// 문자열 유효성 검사
func IsValidUsername(username string) bool {
	// 길이 검증 (최소 2자 이상)
	if len(username) < constants.MinNameLength || len(username) > constants.MaxNameLength {
		return false
	}
	
	// 유니코드 표시 폭 검증 (한글이 2칸 차지하므로 실제 표시 폭 고려)
	displayWidth := GetDisplayWidth(username)
	if displayWidth > 40 { // 한글 20자 또는 영문 40자 정도
		return false
	}
	
	// 빈 문자열이나 공백만 있는 경우 방지
	trimmed := strings.TrimSpace(username)
	if len(trimmed) == 0 {
		return false
	}
	
	// 악의적인 패턴 방지 (보안 강화)
	if containsMaliciousPattern(username) {
		return false
	}
	
	// 특수문자 제한 및 SQL injection, XSS 방지
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9가-힣ㄱ-ㅎ\s\-_.]+$`, username)
	if !matched {
		return false
	}
	
	// 연속된 공백이나 특수문자 방지
	if strings.Contains(username, "  ") || // 연속 공백
		strings.Contains(username, "--") || // 연속 하이픈
		strings.Contains(username, "__") { // 연속 언더스코어
		return false
	}
	
	// 시작/끝이 공백이나 특수문자인 경우 방지
	if len(trimmed) != len(username) || 
		strings.HasPrefix(username, "-") || strings.HasSuffix(username, "-") ||
		strings.HasPrefix(username, "_") || strings.HasSuffix(username, "_") {
		return false
	}
	
	return true
}

// containsMaliciousPattern 악의적인 패턴을 감지합니다
func containsMaliciousPattern(input string) bool {
	// SQL Injection 패턴 감지
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "union", "select", "insert", "update", "delete", "drop", "create", "alter",
		"exec", "execute", "sp_", "xp_", "script", "javascript", "vbscript", "onload", "onerror", "onclick",
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	// 과도한 반복 문자 방지 (DoS 공격 방지)
	if hasExcessiveRepeats(input, 5) {
		return true
	}
	
	// 제어 문자 방지
	for _, char := range input {
		if char < 32 && char != 9 && char != 10 && char != 13 { // 탭, 줄바꿈, 캐리지 리턴 제외
			return true
		}
	}
	
	return false
}

// hasExcessiveRepeats 과도한 문자 반복을 감지합니다
func hasExcessiveRepeats(input string, maxRepeats int) bool {
	if len(input) == 0 {
		return false
	}
	
	count := 1
	prev := rune(0)
	
	for _, char := range input {
		if char == prev {
			count++
			if count > maxRepeats {
				return true
			}
		} else {
			count = 1
			prev = char
		}
	}
	
	return false
}

func IsValidBaekjoonID(id string) bool {
	// 길이 검증
	if len(id) == 0 || len(id) > constants.MaxBaekjoonIDLength || len(id) < constants.MinBaekjoonIDLength {
		return false
	}
	
	// 백준 ID는 영문, 숫자, 언더스코어만 허용
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, id)
	if !matched {
		return false
	}
	
	// 영문으로 시작해야 함
	if !regexp.MustCompile(`^[a-zA-Z]`).MatchString(id) {
		return false
	}
	
	// 연속된 언더스코어 방지
	if strings.Contains(id, "__") {
		return false
	}
	
	// 끝이 언더스코어인 경우 방지
	if strings.HasSuffix(id, "_") {
		return false
	}
	
	return true
}

// 날짜 유효성 검사
func IsValidDateString(dateStr string) bool {
	_, err := time.Parse(constants.DateFormat, dateStr)
	return err == nil
}

func IsValidDateRange(startDate, endDate time.Time) bool {
	return !endDate.Before(startDate)
}

// ParseDateWithValidation 날짜 문자열을 파싱하고 유효성을 검사합니다
func ParseDateWithValidation(dateStr, fieldName string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("%s 날짜가 비어있습니다", fieldName)
	}
	
	parsedDate, err := time.Parse(constants.DateFormat, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s 날짜 형식이 올바르지 않습니다: %s (YYYY-MM-DD 형식으로 입력하세요)", fieldName, dateStr)
	}
	
	return parsedDate, nil
}

// ParseDateRange 시작일과 종료일을 파싱하고 범위를 검증합니다
func ParseDateRange(startDateStr, endDateStr string) (startDate, endDate time.Time, err error) {
	startDate, err = ParseDateWithValidation(startDateStr, "start")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	
	endDate, err = ParseDateWithValidation(endDateStr, "end")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	
	if !IsValidDateRange(startDate, endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("종료일(%s)은 시작일(%s)보다 이전일 수 없습니다", 
			endDateStr, startDateStr)
	}
	
	return startDate, endDate, nil
}

// ValidateAndParseCompetitionDates 대회 생성용 날짜들을 검증하고 파싱합니다
func ValidateAndParseCompetitionDates(name, startDateStr, endDateStr string) (time.Time, time.Time, error) {
	if strings.TrimSpace(name) == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("대회명이 비어있습니다")
	}
	
	return ParseDateRange(startDateStr, endDateStr)
}

// 문자열 처리
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= len(constants.TruncateIndicator) {
		return constants.TruncateIndicator[:maxLen]
	}
	return s[:maxLen-len(constants.TruncateIndicator)] + constants.TruncateIndicator
}

// 한글과 영어 문자 폭을 고려한 문자열 길이 계산
func GetDisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r >= 0x1100 && r <= 0x11FF || // 한글 자모
		   r >= 0x3130 && r <= 0x318F || // 한글 호환 자모
		   r >= 0xAC00 && r <= 0xD7AF || // 한글 완성형
		   r >= 0xFF01 && r <= 0xFF5E {   // 전각 문자
			width += 2 // 한글, 한자 등 전각 문자는 2칸
		} else {
			width += 1 // 영어, 숫자 등 반각 문자는 1칸
		}
	}
	return width
}

func SanitizeString(s string) string {
	// Discord 메시지에서 문제가 될 수 있는 특수문자 제거/변경
	s = strings.ReplaceAll(s, "`", "'")           // 코드 블록 방지
	s = strings.ReplaceAll(s, "<@", "(at)")       // 사용자 멘션 방지 (@ 보다 먼저)
	s = strings.ReplaceAll(s, "<#", "(channel)")  // 채널 멘션 방지
	s = strings.ReplaceAll(s, "<:", "(emoji)")    // 커스텀 이모지 방지
	s = strings.ReplaceAll(s, "@", "(at)")        // 일반 @ 멘션 방지
	s = strings.ReplaceAll(s, "||", "")           // 스포일러 태그 방지
	s = strings.ReplaceAll(s, "**", "")           // 볼드 마크다운 방지
	s = strings.ReplaceAll(s, "__", "")           // 언더라인 마크다운 방지
	s = strings.ReplaceAll(s, "~~", "")           // 취소선 마크다운 방지
	s = strings.ReplaceAll(s, "*", "")            // 이탤릭 마크다운 방지
	
	// 제어 문자 제거
	var cleaned strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\t' { // 출력 가능한 문자만 유지
			cleaned.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(cleaned.String())
}

// SanitizeDiscordMessage Discord 메시지 전용 sanitization
func SanitizeDiscordMessage(s string) string {
	// 기본 sanitization 적용
	s = SanitizeString(s)
	
	// 긴 메시지 자르기 (Discord 메시지 제한: 2000자)
	if len(s) > 1900 { // 여유분 두고 1900자로 제한
		s = s[:1897] + "..."
	}
	
	// 연속된 줄바꿈 제한 (스팸 방지)
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	
	return s
}


