package utils

import (
	"discord-bot/constants"
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

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var globalLogger *Logger

func init() {
	globalLogger = NewLogger()
}

func NewLogger() *Logger {
	level := getLogLevelFromEnv()
	logger := log.New(os.Stdout, "", 0)

	return &Logger{
		level:  level,
		logger: logger,
	}
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

	levelStr := l.getLevelString(level)
	timestamp := time.Now().Format(constants.DateTimeFormat)
	message := fmt.Sprintf(format, args...)

	// 보안: 민감한 정보가 로그에 기록되지 않도록 필터링
	filteredMessage := l.filterSensitiveInfo(message)

	l.logger.Printf("[%s] %s %s", timestamp, levelStr, filteredMessage)
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
