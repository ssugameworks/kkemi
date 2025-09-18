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
	UniversityID = 323 // 숭실대학교 organizationId
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

// 환경 변수 키
const (
	EnvDiscordToken = "DISCORD_BOT_TOKEN"
	EnvChannelID    = "DISCORD_CHANNEL_ID"
	EnvLogLevel     = "LOG_LEVEL"
	EnvDebugMode    = "DEBUG_MODE"
	EnvJSONLogging  = "JSON_LOGGING"
)

// 텔레메트리 관련 상수
const (
	TelemetryNamespace       = "discord-bot"
	TelemetryJobName         = "competition-bot"
	TelemetryTaskID          = "main"
	TelemetryCredentialsFile = "discord-bot-gcloud-credentials.json"
	TelemetryFilePermissions = 0600
)

// 시스템 컴포넌트 이름
const (
	SystemComponentName = "discord-bot"
)

// Google Sheets 관련 상수
const (
	ParticipantSpreadsheetID = "1wwjn1hApSINnYsQGbEe5OdpYWvMfsfHC1ftoyR65IDM"
	ParticipantSheetRange    = "A:Z"           // 전체 시트 범위 (기본 시트)
	ParticipantNameColumn    = "이름\n(ex. 홍길동)" // 실제 스프레드시트 헤더와 정확히 일치
)

// 환경 변수 키 (백업 명단용)
const (
	EnvBackupParticipantList = "BACKUP_PARTICIPANT_LIST"
)

// 에러 메시지 상수
const (
	// 참가자 명단 검증 관련 에러 메시지
	ErrorSheetsCheckFailed = "게임웍스 부원 목록을 불러올 수 없습니다. 잠시 후 다시 시도해주세요."
	ErrorNameNotInList     = "%s님은 게임웍스 부원이 아닙니다.\n만약 부원인 경우에도 이 메시지가 계속 뜬다면 운영진에게 문의해주세요."
)
