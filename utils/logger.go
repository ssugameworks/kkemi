package utils

import (
	"github.com/ssugameworks/Discord-Bot/constants"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// LogEntry 구조화된 로그 엔트리
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Component string                 `json:"component,omitempty"`
}

type Logger struct {
	level     LogLevel
	logger    *log.Logger
	useJSON   bool
	component string
}

var globalLogger *Logger

func init() {
	globalLogger = NewLogger()
}

func NewLogger() *Logger {
	level := getLogLevelFromEnv()
	logger := log.New(os.Stdout, "", 0)
	useJSON := getJSONLoggingFromEnv()

	return &Logger{
		level:     level,
		logger:    logger,
		useJSON:   useJSON,
		component: "discord-bot",
	}
}

// NewComponentLogger 특정 컴포넌트용 로거를 생성합니다
func NewComponentLogger(component string) *Logger {
	level := getLogLevelFromEnv()
	logger := log.New(os.Stdout, "", 0)
	useJSON := getJSONLoggingFromEnv()

	return &Logger{
		level:     level,
		logger:    logger,
		useJSON:   useJSON,
		component: component,
	}
}

// getJSONLoggingFromEnv 환경변수에서 JSON 로깅 설정을 가져옵니다
func getJSONLoggingFromEnv() bool {
	jsonLogging := strings.ToUpper(os.Getenv("JSON_LOGGING"))
	return jsonLogging == "TRUE" || jsonLogging == "1" || jsonLogging == "ON"
}

func getLogLevelFromEnv() LogLevel {
	levelStr := strings.ToUpper(os.Getenv(constants.EnvLogLevel))
	switch levelStr {
	case constants.LogLevelDebug:
		return DEBUG
	case constants.LogLevelInfo:
		return INFO
	case constants.LogLevelWarn:
		return WARN
	case constants.LogLevelError:
		return ERROR
	default:
		return INFO
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	filteredMessage := l.filterSensitiveInfo(message)

	if l.useJSON {
		l.logJSON(level, filteredMessage, nil)
	} else {
		l.logText(level, filteredMessage)
	}
}

// logWithFields 필드와 함께 로그를 기록합니다
func (l *Logger) logWithFields(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	filteredMessage := l.filterSensitiveInfo(message)
	filteredFields := l.filterSensitiveFields(fields)

	if l.useJSON {
		l.logJSON(level, filteredMessage, filteredFields)
	} else {
		l.logText(level, filteredMessage)
	}
}

// logJSON JSON 형태로 로그를 출력합니다
func (l *Logger) logJSON(level LogLevel, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     l.getLevelString(level),
		Message:   message,
		Component: l.component,
		Fields:    fields,
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// JSON 마샬링 실패 시 플레인 텍스트로 폴백
		l.logText(level, fmt.Sprintf("JSON marshal error: %v, original message: %s", err, message))
		return
	}

	l.logger.Println(string(jsonBytes))
}

// logText 텍스트 형태로 로그를 출력합니다
func (l *Logger) logText(level LogLevel, message string) {
	levelStr := l.getLevelString(level)
	timestamp := time.Now().Format(constants.DateTimeFormat)

	if l.component != "" {
		l.logger.Printf("[%s] %s [%s] %s", timestamp, levelStr, l.component, message)
	} else {
		l.logger.Printf("[%s] %s %s", timestamp, levelStr, message)
	}
}

// filterSensitiveInfo 민감한 정보를 로그에서 마스킹합니다
func (l *Logger) filterSensitiveInfo(message string) string {
	// Discord 봇 토큰 패턴 감지 및 마스킹
	if strings.Contains(message, "Bot ") && len(message) > 60 {
		words := strings.Fields(message)
		for i, word := range words {
			if strings.HasPrefix(word, "Bot ") && len(word) > 60 {
				words[i] = "Bot ***TOKEN***"
			} else if len(word) > 50 && strings.Contains(word, ".") {
				words[i] = "***DISCORD_TOKEN***"
			}
		}
		message = strings.Join(words, " ")
	}

	// 기본적인 키워드 기반 마스킹
	sensitiveKeywords := []string{"token", "key", "secret", "password"}
	lowerMessage := strings.ToLower(message)

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerMessage, keyword+"=") ||
			strings.Contains(lowerMessage, keyword+":") ||
			strings.Contains(lowerMessage, keyword+"\"") {
			// 키워드가 포함된 경우 마스킹 처리
			if idx := strings.Index(lowerMessage, keyword); idx != -1 {
				before := message[:idx+len(keyword)]
				remaining := message[idx+len(keyword):]

				// = 또는 : 다음의 값을 찾아 마스킹
				for _, sep := range []string{"=", ":", "\""} {
					if strings.HasPrefix(remaining, sep) {
						message = before + sep + "***MASKED***"
						break
					}
				}
			}
		}
	}

	return message
}

// filterSensitiveFields 필드에서 민감한 정보를 필터링합니다
func (l *Logger) filterSensitiveFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}

	filtered := make(map[string]interface{})
	sensitiveKeys := []string{"token", "password", "secret", "key", "auth", "credential"}

	for k, v := range fields {
		lowerKey := strings.ToLower(k)
		isSensitive := false

		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			filtered[k] = "***MASKED***"
		} else if str, ok := v.(string); ok {
			filtered[k] = l.filterSensitiveInfo(str)
		} else {
			filtered[k] = v
		}
	}

	return filtered
}

func (l *Logger) getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return constants.LogLevelDebug
	case INFO:
		return constants.LogLevelInfo
	case WARN:
		return constants.LogLevelWarn
	case ERROR:
		return constants.LogLevelError
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// WithFields 필드와 함께 로깅하는 메서드들
func (l *Logger) DebugWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(DEBUG, message, fields)
}

func (l *Logger) InfoWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(INFO, message, fields)
}

func (l *Logger) WarnWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(WARN, message, fields)
}

func (l *Logger) ErrorWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(ERROR, message, fields)
}

// 글로벌 로거 함수들
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	globalLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

// 글로벌 필드 로깅 함수들
func DebugWithFields(message string, fields map[string]interface{}) {
	globalLogger.DebugWithFields(message, fields)
}

func InfoWithFields(message string, fields map[string]interface{}) {
	globalLogger.InfoWithFields(message, fields)
}

func WarnWithFields(message string, fields map[string]interface{}) {
	globalLogger.WarnWithFields(message, fields)
}

func ErrorWithFields(message string, fields map[string]interface{}) {
	globalLogger.ErrorWithFields(message, fields)
}
