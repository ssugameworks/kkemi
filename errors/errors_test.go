package errors

import (
	"fmt"
	"github.com/ssugameworks/Discord-Bot/constants"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	code := "TEST_CODE"
	message := "테스트 메시지"
	userMsg := "사용자 메시지"

	err := NewValidationError(code, message, userMsg)

	if err.Type != TypeValidation {
		t.Errorf("Type이 TypeValidation이어야 합니다. 실제값: %v", err.Type)
	}

	if err.Code != code {
		t.Errorf("Code가 %s이어야 합니다. 실제값: %s", code, err.Code)
	}

	if err.Message != message {
		t.Errorf("Message가 %s이어야 합니다. 실제값: %s", message, err.Message)
	}

	if err.UserMsg != userMsg {
		t.Errorf("UserMsg가 %s이어야 합니다. 실제값: %s", userMsg, err.UserMsg)
	}
}

func TestNewAPIError(t *testing.T) {
	code := "API_ERROR"
	message := "API 오류"
	internalErr := fmt.Errorf("내부 오류")

	err := NewAPIError(code, message, internalErr)

	if err.Type != TypeAPI {
		t.Errorf("Type이 TypeAPI여야 합니다. 실제값: %v", err.Type)
	}

	if err.Code != code {
		t.Errorf("Code가 %s이어야 합니다. 실제값: %s", code, err.Code)
	}

	if err.Message != message {
		t.Errorf("Message가 %s이어야 합니다. 실제값: %s", message, err.Message)
	}

	if err.Internal != internalErr {
		t.Errorf("Internal 오류가 설정되지 않음")
	}

	expectedUserMsg := "외부 서비스 연결에 문제가 발생했습니다. 잠시 후 다시 시도해주세요."
	if err.UserMsg != expectedUserMsg {
		t.Errorf("UserMsg가 %s이어야 합니다. 실제값: %s", expectedUserMsg, err.UserMsg)
	}
}

func TestNewNotFoundError(t *testing.T) {
	code := "NOT_FOUND"
	message := "찾을 수 없음"
	userMsg := "리소스를 찾을 수 없습니다"

	err := NewNotFoundError(code, message, userMsg)

	if err.Type != TypeNotFound {
		t.Errorf("Type이 TypeNotFound여야 합니다. 실제값: %v", err.Type)
	}

	if err.Code != code {
		t.Errorf("Code가 %s이어야 합니다. 실제값: %s", code, err.Code)
	}
}

func TestNewDuplicateError(t *testing.T) {
	code := "DUPLICATE"
	message := "중복 리소스"
	userMsg := "이미 존재합니다"

	err := NewDuplicateError(code, message, userMsg)

	if err.Type != TypeDuplicate {
		t.Errorf("Type이 TypeDuplicate여야 합니다. 실제값: %v", err.Type)
	}
}

func TestNewSystemError(t *testing.T) {
	code := "SYSTEM_ERROR"
	message := "시스템 오류"
	internalErr := fmt.Errorf("내부 시스템 오류")

	err := NewSystemError(code, message, internalErr)

	if err.Type != TypeSystem {
		t.Errorf("Type이 TypeSystem이어야 합니다. 실제값: %v", err.Type)
	}

	if err.Internal != internalErr {
		t.Errorf("Internal 오류가 설정되지 않음")
	}

	expectedUserMsg := "시스템 오류가 발생했습니다. 관리자에게 문의해주세요."
	if err.UserMsg != expectedUserMsg {
		t.Errorf("UserMsg가 %s이어야 합니다. 실제값: %s", expectedUserMsg, err.UserMsg)
	}
}

func TestAppErrorError(t *testing.T) {
	// Internal 오류가 있는 경우
	internalErr := fmt.Errorf("내부 오류")
	err := &AppError{
		Code:     "TEST",
		Message:  "테스트 메시지",
		Internal: internalErr,
	}

	expected := "[TEST] 테스트 메시지: 내부 오류"
	if err.Error() != expected {
		t.Errorf("Error() 결과가 %s이어야 합니다. 실제값: %s", expected, err.Error())
	}

	// Internal 오류가 없는 경우
	err2 := &AppError{
		Code:    "TEST2",
		Message: "테스트 메시지2",
	}

	expected2 := "[TEST2] 테스트 메시지2"
	if err2.Error() != expected2 {
		t.Errorf("Error() 결과가 %s이어야 합니다. 실제값: %s", expected2, err2.Error())
	}
}

func TestAppErrorGetUserMessage(t *testing.T) {
	// UserMsg가 있는 경우
	err := &AppError{
		Message: "내부 메시지",
		UserMsg: "사용자 메시지",
	}

	if err.GetUserMessage() != "사용자 메시지" {
		t.Errorf("GetUserMessage()가 '사용자 메시지'를 반환해야 합니다. 실제값: %s", err.GetUserMessage())
	}

	// UserMsg가 없는 경우
	err2 := &AppError{
		Message: "내부 메시지만",
	}

	if err2.GetUserMessage() != "내부 메시지만" {
		t.Errorf("GetUserMessage()가 '내부 메시지만'을 반환해야 합니다. 실제값: %s", err2.GetUserMessage())
	}
}

func TestErrorTypeConstants(t *testing.T) {
	// ErrorType 상수들이 올바른 값을 가지는지 확인
	if TypeValidation != 0 {
		t.Errorf("TypeValidation이 0이어야 합니다. 실제값: %d", TypeValidation)
	}

	if TypeAPI != 1 {
		t.Errorf("TypeAPI가 1이어야 합니다. 실제값: %d", TypeAPI)
	}

	if TypeNotFound != 2 {
		t.Errorf("TypeNotFound가 2여야 합니다. 실제값: %d", TypeNotFound)
	}

	if TypeDuplicate != 3 {
		t.Errorf("TypeDuplicate가 3이어야 합니다. 실제값: %d", TypeDuplicate)
	}

	if TypeSystem != 4 {
		t.Errorf("TypeSystem이 4여야 합니다. 실제값: %d", TypeSystem)
	}
}

// 상수 값 테스트 - 컴파일 타임에 상수들이 올바르게 정의되어 있는지 확인
func TestConstantsExist(t *testing.T) {
	// 이모지 상수들이 존재하는지 확인
	if constants.EmojiSuccess == "" {
		t.Error("EmojiSuccess 상수가 정의되어 있어야 합니다")
	}

	if constants.EmojiError == "" {
		t.Error("EmojiError 상수가 정의되어 있어야 합니다")
	}

	if constants.EmojiInfo == "" {
		t.Error("EmojiInfo 상수가 정의되어 있어야 합니다")
	}

	if constants.EmojiWarning == "" {
		t.Error("EmojiWarning 상수가 정의되어 있어야 합니다")
	}
}
