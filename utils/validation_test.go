package utils

import (
	"github.com/ssugameworks/Discord-Bot/constants"
	"os"
	"testing"
	"time"
)

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		// Valid cases
		{"홍길동", true, "Korean name"},
		{"John Doe", true, "English name with space"},
		{"김철수123", true, "Korean with numbers"},
		{"user_name", true, "English with underscore"},
		{"test-user", true, "English with hyphen"},

		// Invalid cases - length
		{"", false, "empty string"},
		{"a", false, "too short"},
		{"이것은아주긴이름입니다이것은아주긴이름입니다이것은아주긴이름입니다", false, "too long"},

		// Invalid cases - characters
		{"user@domain", false, "contains @ symbol"},
		{"user`name", false, "contains backtick"},
		{"user<>name", false, "contains brackets"},
		{"user||name", false, "contains pipe"},

		// Invalid cases - formatting
		{" leadingspace", false, "leading space"},
		{"trailingspace ", false, "trailing space"},
		{"_underscore", false, "starts with underscore"},
		{"underscore_", false, "ends with underscore"},
		{"-hyphen", false, "starts with hyphen"},
		{"hyphen-", false, "ends with hyphen"},
		{"double  space", false, "double space"},
		{"double--hyphen", false, "double hyphen"},
		{"double__underscore", false, "double underscore"},
	}

	for _, test := range tests {
		result := IsValidUsername(test.input)
		if result != test.expected {
			t.Errorf("IsValidUsername(%q) = %v, expected %v (%s)", test.input, result, test.expected, test.desc)
		}
	}
}

func TestIsValidBaekjoonID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		// Valid cases
		{"alice", true, "simple English"},
		{"user123", true, "English with numbers"},
		{"test_user", true, "English with underscore"},
		{"a1b2c3", true, "mixed alphanumeric"},

		// Invalid cases - length
		{"", false, "empty string"},
		{"ab", false, "too short"},
		{"verylongusernamethatexceedslimit", false, "too long"},

		// Invalid cases - format
		{"123user", false, "starts with number"},
		{"_user", false, "starts with underscore"},
		{"user_", false, "ends with underscore"},
		{"user__name", false, "double underscore"},

		// Invalid cases - characters
		{"user-name", false, "contains hyphen"},
		{"user@domain", false, "contains @ symbol"},
		{"user name", false, "contains space"},
		{"user.name", false, "contains dot"},
		{"한글이름", false, "Korean characters"},
	}

	for _, test := range tests {
		result := IsValidBaekjoonID(test.input)
		if result != test.expected {
			t.Errorf("IsValidBaekjoonID(%q) = %v, expected %v (%s)", test.input, result, test.expected, test.desc)
		}
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"normal text", "normal text", "normal text unchanged"},
		{"text with `code`", "text with 'code'", "backticks to apostrophes"},
		{"@everyone", "(at)everyone", "@ symbol sanitized"},
		{"<@123456>", "(at)123456>", "user mention sanitized"},
		{"<#general>", "(channel)general>", "channel mention sanitized"},
		{"<:emoji:123>", "(emoji)emoji:123>", "custom emoji sanitized"},
		{"**bold** text", "bold text", "markdown bold removed"},
		{"__underline__", "underline", "markdown underline removed"},
		{"~~strikethrough~~", "strikethrough", "markdown strikethrough removed"},
		{"*italic*", "italic", "markdown italic removed"},
		{"||spoiler||", "spoiler", "spoiler tags removed"},
		{"  leading and trailing  ", "leading and trailing", "whitespace trimmed"},
	}

	for _, test := range tests {
		result := SanitizeString(test.input)
		if result != test.expected {
			t.Errorf("SanitizeString(%q) = %q, expected %q (%s)", test.input, result, test.expected, test.desc)
		}
	}
}

func TestGetDisplayWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		desc     string
	}{
		{"hello", 5, "English characters (1 char each)"},
		{"안녕하세요", 10, "Korean characters (2 chars each)"},
		{"hello안녕", 9, "Mixed English and Korean"},
		{"", 0, "empty string"},
		{"123", 3, "numbers"},
	}

	for _, test := range tests {
		result := GetDisplayWidth(test.input)
		if result != test.expected {
			t.Errorf("GetDisplayWidth(%q) = %d, expected %d (%s)", test.input, result, test.expected, test.desc)
		}
	}
}

func TestParseTimeZoneHandling(t *testing.T) {
	// 현재 시간 확인
	now := time.Now()
	t.Logf("Current time: %s", now.Format("2006-01-02 15:04:05 MST"))
	t.Logf("Current UTC time: %s", now.UTC().Format("2006-01-02 15:04:05 MST"))

	// 오늘 날짜로 파싱 테스트
	today := now.Format(constants.DateFormat)
	t.Logf("Today's date string: %s", today)

	parsedDate, err := ParseDateWithValidation(today, "test")
	if err != nil {
		t.Fatalf("Failed to parse today's date: %v", err)
	}

	t.Logf("Parsed date: %s", parsedDate.Format("2006-01-02 15:04:05 MST"))
	t.Logf("Parsed date UTC: %s", parsedDate.UTC().Format("2006-01-02 15:04:05 MST"))

	// 현재 시간이 파싱된 날짜 이후인지 확인
	isAfterStart := !now.Before(parsedDate)
	t.Logf("Is now (%s) after or equal to parsed start (%s)? %v",
		now.Format("2006-01-02 15:04:05 MST"),
		parsedDate.Format("2006-01-02 15:04:05 MST"),
		isAfterStart)

	// 비교 결과 확인
	if now.Hour() >= 0 && !isAfterStart {
		t.Errorf("Expected current time to be after parsed start date, but got false")
		t.Errorf("Current time: %s", now.Format(time.RFC3339))
		t.Errorf("Parsed start: %s", parsedDate.Format(time.RFC3339))
		t.Errorf("Comparison: now.Before(parsedDate) = %v", now.Before(parsedDate))
	}
}

func TestCompetitionDateComparison(t *testing.T) {
	// 2025-09-08로 테스트
	testDateStr := "2025-09-08"

	parsedDate, err := ParseDateWithValidation(testDateStr, "start")
	if err != nil {
		t.Fatalf("Failed to parse test date: %v", err)
	}

	now := GetCurrentTimeKST()

	t.Logf("Test scenario:")
	t.Logf("  Competition start: %s", parsedDate.Format("2006-01-02 15:04:05 MST"))
	t.Logf("  Current time (KST): %s", now.Format("2006-01-02 15:04:05 MST"))
	t.Logf("  now.Before(start): %v", now.Before(parsedDate))

	// 현재가 9월 8일 이후라면 등록 가능해야 함
	if now.Year() == 2025 && now.Month() == 9 && now.Day() >= 8 {
		if now.Before(parsedDate) {
			t.Errorf("Expected registration to be allowed on 2025-09-08, but now.Before(start) = true")
			t.Errorf("This suggests timezone handling is incorrect")
		}
	}
}

func TestTimeZoneConsistency(t *testing.T) {
	// 시간대 일관성 테스트
	testDate := "2025-09-08"

	parsed, err := ParseDateWithValidation(testDate, "test")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	now := GetCurrentTimeKST()

	t.Logf("Timezone consistency check:")
	t.Logf("  Parsed date location: %s", parsed.Location())
	t.Logf("  Current time location: %s", now.Location())
	t.Logf("  Are they the same timezone? %v", parsed.Location() == now.Location())

	if parsed.Location().String() != now.Location().String() {
		t.Errorf("Timezone mismatch! Parsed date and current time should be in same timezone")
	}
}

func TestIsNameInBackupList(t *testing.T) {
	// 테스트용 환경변수 설정
	originalEnv := os.Getenv(constants.EnvBackupParticipantList)
	defer func() {
		if originalEnv != "" {
			os.Setenv(constants.EnvBackupParticipantList, originalEnv)
		} else {
			os.Unsetenv(constants.EnvBackupParticipantList)
		}
	}()

	// 테스트 케이스 1: 환경변수가 설정된 경우
	os.Setenv(constants.EnvBackupParticipantList, "공서연, 이정안, 김철수")
	
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"공서연", true, "백업 명단에 있는 이름"},
		{"이정안", true, "백업 명단에 있는 다른 이름"},
		{"김철수", true, "백업 명단에 있는 세 번째 이름"},
		{"공 서 연", true, "공백이 있는 이름 (정규화 후 일치)"},
		{"UNKNOWN", false, "백업 명단에 없는 이름"},
		{"", false, "빈 문자열"},
		{" 공서연 ", true, "앞뒤 공백이 있는 이름"},
	}

	for _, test := range tests {
		result := IsNameInBackupList(test.input)
		if result != test.expected {
			t.Errorf("IsNameInBackupList(%q) = %v, expected %v (%s)", test.input, result, test.expected, test.desc)
		}
	}
	
	// 테스트 케이스 2: 환경변수가 없는 경우
	os.Unsetenv(constants.EnvBackupParticipantList)
	
	result := IsNameInBackupList("공서연")
	if result != false {
		t.Errorf("IsNameInBackupList with no env var should return false, got %v", result)
	}
	
	// 백업 명단 내용 확인
	t.Logf("Backup participant list from env: %v", getBackupParticipantList())
}

func TestNormalizeNameForComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"홍길동", "홍길동", "일반적인 한국 이름"},
		{" 홍길동 ", "홍길동", "앞뒤 공백 제거"},
		{"홍 길 동", "홍길동", "중간 공백 제거"},
		{"John Doe", "johndoe", "영어 이름 소문자 변환 및 공백 제거"},
		{"  JANE  SMITH  ", "janesmith", "복합 공백 및 대소문자 처리"},
		{"", "", "빈 문자열"},
	}

	for _, test := range tests {
		result := normalizeNameForComparison(test.input)
		if result != test.expected {
			t.Errorf("normalizeNameForComparison(%q) = %q, expected %q (%s)", test.input, result, test.expected, test.desc)
		}
	}
}
