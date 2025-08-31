package constants

import "time"

// 파일 관련 상수
const (
	ParticipantsFileName = "participants.json"
	CompetitionFileName  = "competition.json"
	FilePermission       = 0644
	BackupFileSuffix     = ".corrupted"
	JSONIndentSpaces     = "  "
)

// API 관련 상수
const (
	SolvedACBaseURL       = "https://solved.ac/api/v3"
	APITimeout            = 30 * time.Second
	MaxRetries            = 3
	RetryDelay            = 1 * time.Second
	APIRetryMultiplier    = 2
	MaxConcurrentRequests = 5
)

// 점수 계산 상수
const (
	ChallengeMultiplier = 1.4
	BaseMultiplier      = 1.0
	PenaltyMultiplier   = 0.5
)

// 대회 관련 상수
const (
	BlackoutDays          = 3
	DailyScoreboardHour   = 9
	DailyScoreboardMinute = 0
	SchedulerInterval     = 24 * time.Hour
	SchedulerTimeout      = 30 * time.Second
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
	MaxUsernameLength    = 15
	TruncateIndicator    = "..."
	ScoreboardRankWidth  = 4
	ScoreboardNameWidth  = 15
	ScoreboardScoreWidth = 6
	ScoreboardSeparator  = "──────────────────────────────"
)

// 메시지 템플릿
const (
	DMReceivedTemplate  = "DM 수신: %s from %s\n"
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
)
