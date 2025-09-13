package constants

import "time"

// API 관련 상수
const (
	SolvedACBaseURL       = "https://solved.ac/api/v3"
	APITimeout            = 30 * time.Second
	MaxRetries            = 3
	RetryDelay            = 1 * time.Second
	APIRetryMultiplier    = 2
	MaxConcurrentRequests = 5
)

// 조직 ID 관련 상수
const (
	SoongsilUniversityID = 323 // 숭실대학교 solved.ac organizationId
)

// 잔디심기 챌린지 리그 분류 상수
const (
	LeagueRookie = 0 // 루키: Unrated ~ Silver V (티어 0-6)
	LeaguePro    = 1 // 프로: Silver IV ~ Gold V (티어 7-11)
	LeagueMax    = 2 // 맥스: Gold IV ~ (티어 12 이상)
)

// 각 리그별 가중치 (상위/동일/하위 티어)
const (
	// 루키 리그 가중치
	RookieUpperMultiplier = 1.4 // 상위 티어 문제
	RookieBaseMultiplier  = 1.0 // 동일 티어 문제
	RookieLowerMultiplier = 0.5 // 하위 티어 문제

	// 프로 리그 가중치
	ProUpperMultiplier = 1.2 // 상위 티어 문제
	ProBaseMultiplier  = 1.0 // 동일 티어 문제
	ProLowerMultiplier = 0.8 // 하위 티어 문제

	// 맥스 리그 가중치
	MaxUpperMultiplier = 1.0 // 상위 티어 문제
	MaxBaseMultiplier  = 1.0 // 동일 티어 문제
	MaxLowerMultiplier = 1.0 // 하위 티어 문제
)

// 대회 관련 상수
const (
	BlackoutDays          = 3
	DailyScoreboardHour   = 9
	DailyScoreboardMinute = 0
	SchedulerInterval     = 24 * time.Hour
)

// Discord 관련 상수
const (
	CommandPrefix = "!"
)

// 이모지 상수
const (
	EmojiSuccess  = "✅"
	EmojiError    = "❌"
	EmojiInfo     = "ℹ️"
	EmojiWarning  = "⚠️"
	EmojiTrophy   = "🏆"
	EmojiUser     = "👤"
	EmojiTarget   = "🎯"
	EmojiMedal    = "🏅"
	EmojiStats    = "📊"
	EmojiCalendar = "📅"
	EmojiClock    = "⏰"
	EmojiLock     = "🔒"
	EmojiPeople   = "👥"
)

// 날짜 형식
const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
)

// 로그 관련 상수
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

// 문자열 크기 제한
const (
	TruncateIndicator    = "..."
	ScoreboardRankWidth  = 4
	ScoreboardNameWidth  = 15
	ScoreboardScoreWidth = 6
	ScoreboardSeparator  = "──────────────────────────────"
)

// 메시지 템플릿
const (
	CommandPrefixLength = 1 // "!" 길이
)

// 티어별 색상 (deprecated - use models.TierManager instead)
const (
	ColorTierGold = 0xE09E37 // 골드 - 스코어보드 기본 색상용
)

// ANSI 색상 코드 (deprecated - use models.TierManager instead)
const (
	ANSIReset = "\x1b[0m"
)

// 환경 변수 키
const (
	EnvDiscordToken = "DISCORD_BOT_TOKEN"
	EnvChannelID    = "DISCORD_CHANNEL_ID"
	EnvLogLevel     = "LOG_LEVEL"
	EnvDebugMode    = "DEBUG_MODE"
	EnvJSONLogging  = "JSON_LOGGING"
)
