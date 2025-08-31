package utils

import (
	"strings"
	"testing"
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

func TestSanitizeDiscordMessage(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		desc   string
	}{
		{"normal message", 14, "normal message unchanged"},
		{"message\n\n\nwith\n\n\nmany\n\n\nlinebreaks", 0, "multiple linebreaks reduced"},
	}

	for _, test := range tests {
		result := SanitizeDiscordMessage(test.input)
		if test.maxLen > 0 && len(result) != test.maxLen {
			t.Errorf("SanitizeDiscordMessage(%q) length = %d, expected %d (%s)", test.input, len(result), test.maxLen, test.desc)
		}
		if test.desc == "multiple linebreaks reduced" && strings.Contains(result, "\n\n\n") {
			t.Errorf("SanitizeDiscordMessage should reduce multiple linebreaks, got %q", result)
		}
	}

	// Test long message truncation
	longMessage := strings.Repeat("a", 2000)
	result := SanitizeDiscordMessage(longMessage)
	if len(result) > 1900 {
		t.Errorf("SanitizeDiscordMessage should truncate long messages to 1900 chars, got %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Errorf("SanitizeDiscordMessage should add ellipsis to truncated messages")
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
