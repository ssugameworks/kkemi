package config

import (
	"github.com/ssugameworks/kkemi/constants"
	"os"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	// 기본 유효한 설정
	validConfig := &Config{
		Discord: DiscordConfig{
			Token:     "valid_token",
			ChannelID: "123456789",
		},
		Schedule: ScheduleConfig{
			ScoreboardHour:   12,
			ScoreboardMinute: 30,
			Enabled:          true,
		},
		Logging: LoggingConfig{
			Level:     constants.LogLevelInfo,
			DebugMode: false,
		},
	}

	// 유효한 설정 테스트
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}

	// Discord 토큰 누락 테스트
	invalidTokenConfig := *validConfig
	invalidTokenConfig.Discord.Token = ""
	if err := invalidTokenConfig.Validate(); err == nil {
		t.Error("Config with empty token should return error")
	}

	// 잘못된 로그 레벨 테스트
	invalidLogLevelConfig := *validConfig
	invalidLogLevelConfig.Logging.Level = "INVALID_LEVEL"
	if err := invalidLogLevelConfig.Validate(); err == nil {
		t.Error("Config with invalid log level should return error")
	}

	// 잘못된 스케줄 시간 테스트 (25시)
	invalidHourConfig := *validConfig
	invalidHourConfig.Schedule.ScoreboardHour = 25
	if err := invalidHourConfig.Validate(); err == nil {
		t.Error("Config with invalid hour (25) should return error")
	}

	// 잘못된 스케줄 분 테스트 (60분)
	invalidMinuteConfig := *validConfig
	invalidMinuteConfig.Schedule.ScoreboardMinute = 60
	if err := invalidMinuteConfig.Validate(); err == nil {
		t.Error("Config with invalid minute (60) should return error")
	}

	// 스케줄이 비활성화된 경우 시간 검증 건너뛰기
	disabledScheduleConfig := *validConfig
	disabledScheduleConfig.Schedule.Enabled = false
	disabledScheduleConfig.Schedule.ScoreboardHour = 25 // 잘못된 값이지만 비활성화되어 있음
	if err := disabledScheduleConfig.Validate(); err != nil {
		t.Error("Config with disabled schedule should not validate schedule times")
	}
}

func TestValidLogLevels(t *testing.T) {
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "debug", "info", "warn", "error"}

	baseConfig := &Config{
		Discord: DiscordConfig{Token: "test_token"},
		Schedule: ScheduleConfig{
			ScoreboardHour:   12,
			ScoreboardMinute: 30,
			Enabled:          false, // 스케줄 검증은 건너뛰기
		},
	}

	for _, level := range validLevels {
		config := *baseConfig
		config.Logging.Level = level
		if err := config.Validate(); err != nil {
			t.Errorf("Log level '%s' should be valid but got error: %v", level, err)
		}
	}
}

func TestBoundaryValues(t *testing.T) {
	baseConfig := &Config{
		Discord:  DiscordConfig{Token: "test_token"},
		Logging:  LoggingConfig{Level: constants.LogLevelInfo},
		Schedule: ScheduleConfig{Enabled: true},
	}

	// 경계값 테스트 - 유효한 값들
	validCombinations := []struct {
		hour   int
		minute int
	}{
		{0, 0},   // 최소값
		{23, 59}, // 최대값
		{12, 30}, // 중간값
	}

	for _, combo := range validCombinations {
		config := *baseConfig
		config.Schedule.ScoreboardHour = combo.hour
		config.Schedule.ScoreboardMinute = combo.minute
		if err := config.Validate(); err != nil {
			t.Errorf("Valid time %02d:%02d should not return error: %v", combo.hour, combo.minute, err)
		}
	}

	// 경계값 테스트 - 무효한 값들
	invalidCombinations := []struct {
		hour   int
		minute int
	}{
		{-1, 0}, // 음수 시간
		{24, 0}, // 24시
		{0, -1}, // 음수 분
		{0, 60}, // 60분
	}

	for _, combo := range invalidCombinations {
		config := *baseConfig
		config.Schedule.ScoreboardHour = combo.hour
		config.Schedule.ScoreboardMinute = combo.minute
		if err := config.Validate(); err == nil {
			t.Errorf("Invalid time %02d:%02d should return error", combo.hour, combo.minute)
		}
	}
}

// 환경변수를 통한 설정 로드 테스트
func TestLoadFromEnv(t *testing.T) {
	// 환경변수 설정
	os.Setenv("DISCORD_BOT_TOKEN", "test_token")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("SCOREBOARD_HOUR", "15")
	os.Setenv("SCOREBOARD_MINUTE", "45")

	// 테스트 후 정리
	defer func() {
		os.Unsetenv("DISCORD_BOT_TOKEN")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SCOREBOARD_HOUR")
		os.Unsetenv("SCOREBOARD_MINUTE")
	}()

	config := Load()

	if config.Discord.Token != "test_token" {
		t.Errorf("Expected token 'test_token', got '%s'", config.Discord.Token)
	}

	if config.Logging.Level != "DEBUG" {
		t.Errorf("Expected log level 'DEBUG', got '%s'", config.Logging.Level)
	}

	if config.Schedule.ScoreboardHour != 15 {
		t.Errorf("Expected hour 15, got %d", config.Schedule.ScoreboardHour)
	}

	if config.Schedule.ScoreboardMinute != 45 {
		t.Errorf("Expected minute 45, got %d", config.Schedule.ScoreboardMinute)
	}

	// 로드된 설정이 유효한지 확인
	if err := config.Validate(); err != nil {
		t.Errorf("Loaded config should be valid: %v", err)
	}
}
