package config

import (
	"discord-bot/constants"
	"os"
	"strconv"
	"strings"
)

// Config 애플리케이션의 전체 설정을 관리합니다
type Config struct {
	Discord   DiscordConfig
	Schedule  ScheduleConfig
	Logging   LoggingConfig
	Features  FeatureFlags
	Telemetry TelemetryConfig
}

type DiscordConfig struct {
	Token     string
	ChannelID string
}

type ScheduleConfig struct {
	ScoreboardHour   int
	ScoreboardMinute int
	Enabled          bool
}

type LoggingConfig struct {
	Level     string
	DebugMode bool
}

type FeatureFlags struct {
	EnableAutoScoreboard bool
	EnableDetailedErrors bool
}

type TelemetryConfig struct {
	Enabled   bool
	ProjectID string
}

// Load는 환경변수에서 설정을 로드합니다
func Load() *Config {
	return &Config{
		Discord: DiscordConfig{
			Token:     getEnv(constants.EnvDiscordToken, ""),
			ChannelID: getEnv(constants.EnvChannelID, ""),
		},
		Schedule: ScheduleConfig{
			ScoreboardHour:   getEnvInt("SCOREBOARD_HOUR", constants.DailyScoreboardHour),
			ScoreboardMinute: getEnvInt("SCOREBOARD_MINUTE", constants.DailyScoreboardMinute),
			Enabled:          getEnv(constants.EnvChannelID, "") != "",
		},
		Logging: LoggingConfig{
			Level:     getEnv(constants.EnvLogLevel, constants.LogLevelInfo),
			DebugMode: getEnvBool(constants.EnvDebugMode, false),
		},
		Features: FeatureFlags{
			EnableAutoScoreboard: getEnvBool("ENABLE_AUTO_SCOREBOARD", true),
			EnableDetailedErrors: getEnvBool("ENABLE_DETAILED_ERRORS", false),
		},
		Telemetry: TelemetryConfig{
			Enabled:   getEnvBool("TELEMETRY_ENABLED", false),
			ProjectID: getEnv("GOOGLE_CLOUD_PROJECT", ""),
		},
	}
}

// Validate 설정의 유효성을 검사합니다
func (c *Config) Validate() error {
	// Discord 설정 검증
	if c.Discord.Token == "" {
		return &ConfigError{
			Field:   "Discord.Token",
			Message: "Discord bot token is required",
		}
	}

	// 로그 레벨 검증
	validLogLevels := map[string]bool{
		constants.LogLevelDebug: true,
		constants.LogLevelInfo:  true,
		constants.LogLevelWarn:  true,
		constants.LogLevelError: true,
	}
	if !validLogLevels[strings.ToUpper(c.Logging.Level)] {
		return &ConfigError{
			Field:   "Logging.Level",
			Message: "LOG_LEVEL must be one of: DEBUG, INFO, WARN, ERROR (got: " + c.Logging.Level + ")",
		}
	}

	// 스케줄 설정 검증 (활성화된 경우에만)
	if c.Schedule.Enabled {
		if c.Schedule.ScoreboardHour < 0 || c.Schedule.ScoreboardHour > 23 {
			return &ConfigError{
				Field:   "Schedule.ScoreboardHour",
				Message: "SCOREBOARD_HOUR must be between 0 and 23 (got: " + strconv.Itoa(c.Schedule.ScoreboardHour) + ")",
			}
		}

		if c.Schedule.ScoreboardMinute < 0 || c.Schedule.ScoreboardMinute > 59 {
			return &ConfigError{
				Field:   "Schedule.ScoreboardMinute",
				Message: "SCOREBOARD_MINUTE must be between 0 and 59 (got: " + strconv.Itoa(c.Schedule.ScoreboardMinute) + ")",
			}
		}
	}

	return nil
}

// IsDebugMode 디버그 모드 여부를 반환합니다
func (c *Config) IsDebugMode() bool {
	return c.Logging.DebugMode || strings.ToUpper(c.Logging.Level) == constants.LogLevelDebug
}

// ConfigError 설정 관련 오류를 나타냅니다
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error in " + e.Field + ": " + e.Message
}

// 헬퍼 함수들
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
