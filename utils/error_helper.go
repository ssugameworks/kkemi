package utils

import (
	"fmt"
	"log"
	"strings"
)

// ErrorHelper 일관된 에러 처리를 위한 헬퍼 구조체입니다
type ErrorHelper struct {
	operation string
}

// NewErrorHelper 새로운 ErrorHelper를 생성합니다
func NewErrorHelper(operation string) *ErrorHelper {
	return &ErrorHelper{
		operation: operation,
	}
}

// WrapError 기존 에러에 컨텍스트를 추가하여 래핑합니다
func (e *ErrorHelper) WrapError(err error, message string) error {
	if err == nil {
		return nil
	}

	if message != "" {
		return fmt.Errorf("[%s] %s: %w", e.operation, message, err)
	}
	return fmt.Errorf("[%s]: %w", e.operation, err)
}

// CreateError 새로운 에러를 생성합니다
func (e *ErrorHelper) CreateError(message string) error {
	return fmt.Errorf("[%s] %s", e.operation, message)
}

// CreateErrorf 포맷된 에러를 생성합니다
func (e *ErrorHelper) CreateErrorf(format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("[%s] %s", e.operation, message)
}

// LogError 에러를 로그에 출력합니다
func (e *ErrorHelper) LogError(err error, context string) {
	if err == nil {
		return
	}

	if context != "" {
		log.Printf("[ERROR] %s - %s: %v", e.operation, context, err)
	} else {
		log.Printf("[ERROR] %s: %v", e.operation, err)
	}
}

// IsNotFoundError 에러가 "not found" 타입인지 확인합니다
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "찾을 수 없습니다")
}

// IsValidationError 에러가 유효성 검증 에러인지 확인합니다
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "validation") ||
		strings.Contains(errMsg, "올바르지 않습니다") || strings.Contains(errMsg, "형식")
}

// IsAPIError 에러가 API 관련 에러인지 확인합니다
func IsAPIError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "api") || strings.Contains(errMsg, "http") ||
		strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "request") || strings.Contains(errMsg, "response")
}

// GetUserFriendlyMessage 에러에 따라 사용자 친화적인 메시지를 반환합니다
func GetUserFriendlyMessage(err error) string {
	if err == nil {
		return "알 수 없는 오류가 발생했습니다."
	}

	if IsNotFoundError(err) {
		return "요청한 정보를 찾을 수 없습니다."
	}

	if IsValidationError(err) {
		return "입력한 정보의 형식이 올바르지 않습니다."
	}

	if IsAPIError(err) {
		return "외부 서비스와의 통신 중 문제가 발생했습니다."
	}

	return "처리 중 오류가 발생했습니다. 잠시 후 다시 시도해주세요."
}
