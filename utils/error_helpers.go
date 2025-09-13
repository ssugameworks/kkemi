package utils

import (
	"discord-bot/constants"
	"discord-bot/errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ValidationErrorHelper 검증 에러 처리를 위한 헬퍼
type ValidationErrorHelper struct {
	session   *discordgo.Session
	channelID string
}

// NewValidationErrorHelper ValidationErrorHelper 생성자
func NewValidationErrorHelper(session *discordgo.Session, channelID string) *ValidationErrorHelper {
	return &ValidationErrorHelper{
		session:   session,
		channelID: channelID,
	}
}

// HandleInvalidParams 잘못된 매개변수 에러 처리
func (v *ValidationErrorHelper) HandleInvalidParams(code, message, userMsg string) {
	err := errors.NewValidationError(code, message, userMsg)
	errors.HandleDiscordError(v.session, v.channelID, err)
}

// HandleInsufficientPermissions 권한 부족 에러 처리
func (v *ValidationErrorHelper) HandleInsufficientPermissions() {
	err := errors.NewValidationError("INSUFFICIENT_PERMISSIONS",
		"사용자가 필수 권한을 가지고 있지 않습니다",
		constants.MsgInsufficientPermissions)
	errors.HandleDiscordError(v.session, v.channelID, err)
}

// HandleInvalidDateFormat 잘못된 날짜 형식 에러 처리
func (v *ValidationErrorHelper) HandleInvalidDateFormat(field string) {
	err := errors.NewValidationError(fmt.Sprintf("INVALID_%s_DATE", field),
		fmt.Sprintf("%s 날짜 형식이 올바르지 않습니다", field),
		"날짜 형식이 올바르지 않습니다. (YYYY-MM-DD)")
	errors.HandleDiscordError(v.session, v.channelID, err)
}

// HandleInvalidDateRange 잘못된 날짜 범위 에러 처리
func (v *ValidationErrorHelper) HandleInvalidDateRange() {
	err := errors.NewValidationError("INVALID_DATE_RANGE",
		"종료일은 시작일보다 이전일 수 없습니다",
		"종료일은 시작일보다 빨라질 수 없습니다.")
	errors.HandleDiscordError(v.session, v.channelID, err)
}

// SystemErrorHelper 시스템 에러 처리를 위한 헬퍼
type SystemErrorHelper struct {
	session   *discordgo.Session
	channelID string
}

// NewSystemErrorHelper SystemErrorHelper 생성자
func NewSystemErrorHelper(session *discordgo.Session, channelID string) *SystemErrorHelper {
	return &SystemErrorHelper{
		session:   session,
		channelID: channelID,
	}
}

// HandleSystemError 시스템 에러 처리
func (s *SystemErrorHelper) HandleSystemError(code, message, userMsg string, err error) {
	botErr := errors.NewSystemError(code, message, err)
	botErr.UserMsg = userMsg
	errors.HandleDiscordError(s.session, s.channelID, botErr)
}

// HandleCompetitionCreateFailed 대회 생성 실패 에러 처리
func (s *SystemErrorHelper) HandleCompetitionCreateFailed(err error) {
	botErr := errors.NewSystemError("COMPETITION_CREATE_FAILED",
		"대회 생성에 실패했습니다", err)
	botErr.UserMsg = "대회 생성에 실패했습니다."
	errors.HandleDiscordError(s.session, s.channelID, botErr)
}

// HandleCompetitionUpdateFailed 대회 업데이트 실패 에러 처리
func (s *SystemErrorHelper) HandleCompetitionUpdateFailed(err error) {
	botErr := errors.NewSystemError("COMPETITION_UPDATE_FAILED",
		"대회 수정에 실패했습니다", err)
	botErr.UserMsg = "대회 정보 업데이트에 실패했습니다."
	errors.HandleDiscordError(s.session, s.channelID, botErr)
}

// HandleScoreboardGenerationFailed 스코어보드 생성 실패 에러 처리
func (s *SystemErrorHelper) HandleScoreboardGenerationFailed(err error) {
	botErr := errors.NewSystemError("SCOREBOARD_GENERATION_FAILED",
		"스코어보드 생성에 실패했습니다", err)
	botErr.UserMsg = "스코어보드 생성에 실패했습니다."
	errors.HandleDiscordError(s.session, s.channelID, botErr)
}

// APIErrorHelper API 에러 처리를 위한 헬퍼
type APIErrorHelper struct {
	session   *discordgo.Session
	channelID string
}

// NewAPIErrorHelper APIErrorHelper 생성자
func NewAPIErrorHelper(session *discordgo.Session, channelID string) *APIErrorHelper {
	return &APIErrorHelper{
		session:   session,
		channelID: channelID,
	}
}

// HandleBaekjoonUserNotFound 백준 사용자 찾기 실패 에러 처리
func (a *APIErrorHelper) HandleBaekjoonUserNotFound(baekjoonID string, err error) {
	botErr := errors.NewAPIError("BAEKJOON_USER_NOT_FOUND",
		fmt.Sprintf("백준 사용자 '%s'를 찾을 수 없습니다", baekjoonID), err)
	botErr.UserMsg = fmt.Sprintf("백준 사용자 '%s'를 찾을 수 없습니다.", baekjoonID)
	errors.HandleDiscordError(a.session, a.channelID, botErr)
}

// DataErrorHelper 데이터 관련 에러 처리를 위한 헬퍼
type DataErrorHelper struct {
	session   *discordgo.Session
	channelID string
}

// NewDataErrorHelper DataErrorHelper 생성자
func NewDataErrorHelper(session *discordgo.Session, channelID string) *DataErrorHelper {
	return &DataErrorHelper{
		session:   session,
		channelID: channelID,
	}
}

// HandleParticipantAlreadyExists 참가자 중복 등록 에러 처리
func (d *DataErrorHelper) HandleParticipantAlreadyExists(baekjoonID string) {
	botErr := errors.NewDuplicateError("PARTICIPANT_ALREADY_EXISTS",
		fmt.Sprintf("백준 ID '%s'로 이미 등록된 참가자가 있습니다", baekjoonID),
		fmt.Sprintf("백준 ID '%s'로 이미 등록된 참가자가 있습니다.", baekjoonID))
	errors.HandleDiscordError(d.session, d.channelID, botErr)
}

// HandleParticipantNotFound 참가자 찾기 실패 에러 처리
func (d *DataErrorHelper) HandleParticipantNotFound(baekjoonID string) {
	botErr := errors.NewNotFoundError("PARTICIPANT_NOT_FOUND",
		fmt.Sprintf("백준 ID '%s'로 등록된 참가자를 찾을 수 없습니다", baekjoonID),
		fmt.Sprintf("백준 ID '%s'로 등록된 참가자를 찾을 수 없습니다.", baekjoonID))
	errors.HandleDiscordError(d.session, d.channelID, botErr)
}

// HandleNoActiveCompetition 활성 대회 없음 에러 처리
func (d *DataErrorHelper) HandleNoActiveCompetition() {
	err := errors.NewNotFoundError("NO_ACTIVE_COMPETITION",
		"활성화된 대회를 찾을 수 없습니다",
		"현재 진행 중인 대회가 없습니다.")
	errors.HandleDiscordError(d.session, d.channelID, err)
}

// ErrorHandlerFactory 에러 핸들러들을 생성하는 팩토리
type ErrorHandlerFactory struct {
	session   *discordgo.Session
	channelID string
}

// NewErrorHandlerFactory ErrorHandlerFactory 생성자
func NewErrorHandlerFactory(session *discordgo.Session, channelID string) *ErrorHandlerFactory {
	return &ErrorHandlerFactory{
		session:   session,
		channelID: channelID,
	}
}

// Validation ValidationErrorHelper 반환
func (f *ErrorHandlerFactory) Validation() *ValidationErrorHelper {
	return NewValidationErrorHelper(f.session, f.channelID)
}

// System SystemErrorHelper 반환
func (f *ErrorHandlerFactory) System() *SystemErrorHelper {
	return NewSystemErrorHelper(f.session, f.channelID)
}

// API APIErrorHelper 반환
func (f *ErrorHandlerFactory) API() *APIErrorHelper {
	return NewAPIErrorHelper(f.session, f.channelID)
}

// Data DataErrorHelper 반환
func (f *ErrorHandlerFactory) Data() *DataErrorHelper {
	return NewDataErrorHelper(f.session, f.channelID)
}
