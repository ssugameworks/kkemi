package errors

import (
	"discord-bot/constants"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// ErrorType 오류의 종류를 나타냅니다
type ErrorType int

const (
	TypeValidation ErrorType = iota
	TypeAPI
	TypeNotFound
	TypeDuplicate
	TypeSystem
)

// AppError 애플리케이션에서 발생하는 구조화된 오류를 표현합니다
type AppError struct {
	Type     ErrorType
	Code     string
	Message  string
	UserMsg  string
	Internal error
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// GetUserMessage 사용자에게 표시할 메시지를 반환합니다
func (e *AppError) GetUserMessage() string {
	if e.UserMsg != "" {
		return e.UserMsg
	}
	return e.Message
}

// 오류 생성 함수들

// NewValidationError 입력값 검증 오류를 생성합니다
func NewValidationError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypeValidation,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewAPIError 외부 API 연동 오류를 생성합니다
func NewAPIError(code, message string, err error) *AppError {
	return &AppError{
		Type:     TypeAPI,
		Code:     code,
		Message:  message,
		UserMsg:  "외부 서비스 연결에 문제가 발생했습니다. 잠시 후 다시 시도해주세요.",
		Internal: err,
	}
}

// NewNotFoundError 리소스를 찾을 수 없는 오류를 생성합니다
func NewNotFoundError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypeNotFound,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewDuplicateError 중복 리소스 오류를 생성합니다
func NewDuplicateError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypeDuplicate,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewSystemError 시스템 내부 오류를 생성합니다
func NewSystemError(code, message string, err error) *AppError {
	return &AppError{
		Type:     TypeSystem,
		Code:     code,
		Message:  message,
		UserMsg:  "시스템 오류가 발생했습니다. 관리자에게 문의해주세요.",
		Internal: err,
	}
}

// Discord 메시지 관련 헬퍼 함수들

// HandleDiscordError 오류를 처리하고 Discord 채널에 메시지를 전송합니다
func HandleDiscordError(s *discordgo.Session, channelID string, err error) {
	// 이 함수를 호출하기 전에 상세 오류는 이미 로깅되었다고 가정함.
	// 이 함수의 주 목적은 사용자에게 피드백을 주는 것.
	var userMessage string
	if appErr, ok := err.(*AppError); ok {
		userMessage = appErr.GetUserMessage()
	} else {
		userMessage = "예상치 못한 오류가 발생했습니다."
	}

	if discordErr := SendDiscordMessageWithRetry(s, channelID, constants.EmojiError+" "+userMessage); discordErr != nil {
		// 디스코드 메시지 전송 실패는 콘솔에 기록 (최후의 보루)
		// utils import 순환 참조 방지를 위해 fmt 사용
		fmt.Printf("FATAL: Failed to send error message to Discord. ChannelID: %s, OriginalError: %v, DiscordError: %v\n", channelID, err, discordErr)
	}
}

// SendDiscordSuccess 성공 메시지를 Discord 채널에 전송합니다
func SendDiscordSuccess(s *discordgo.Session, channelID, message string) error {
	return SendDiscordMessageWithRetry(s, channelID, constants.EmojiSuccess+" "+message)
}

// SendDiscordInfo 정보 메시지를 Discord 채널에 전송합니다
func SendDiscordInfo(s *discordgo.Session, channelID, message string) error {
	return SendDiscordMessageWithRetry(s, channelID, constants.EmojiInfo+" "+message)
}

// SendDiscordWarning 경고 메시지를 Discord 채널에 전송합니다
func SendDiscordWarning(s *discordgo.Session, channelID, message string) error {
	return SendDiscordMessageWithRetry(s, channelID, constants.EmojiWarning+" "+message)
}

// SendDiscordMessageWithRetry Discord 메시지 전송을 재시도 로직과 함께 수행합니다
func SendDiscordMessageWithRetry(s *discordgo.Session, channelID, message string) error {
	const maxRetries = constants.MaxDiscordRetries
	const baseDelay = constants.BaseRetryDelay

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := s.ChannelMessageSend(channelID, message)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxRetries-1 {
			delay := time.Duration(1<<attempt) * baseDelay // Exponential backoff: 1s, 2s, 4s
			time.Sleep(delay)
		}
	}

	return lastErr
}
