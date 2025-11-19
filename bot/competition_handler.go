package bot

import (
	"fmt"
	"strings"

	"github.com/ssugameworks/kkemi/constants"
	"github.com/ssugameworks/kkemi/errors"
	"github.com/ssugameworks/kkemi/models"
	"github.com/ssugameworks/kkemi/utils"

	"github.com/bwmarrin/discordgo"
)

// CompetitionHandler 대회 관련 명령어를 처리합니다
type CompetitionHandler struct {
	commandHandler *CommandHandler
}

// NewCompetitionHandler 새로운 CompetitionHandler 인스턴스를 생성합니다
func NewCompetitionHandler(ch *CommandHandler) *CompetitionHandler {
	return &CompetitionHandler{
		commandHandler: ch,
	}
}

// HandleCompetition 대회 관련 명령어를 처리합니다
func (ch *CompetitionHandler) HandleCompetition(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// DM이 아닌 경우에만 관리자 권한 확인
	if m.GuildID != "" && !ch.commandHandler.isAdmin(s, m) {
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	if len(params) == 0 {
		errorHandlers.Validation().HandleInvalidParams("COMPETITION_INVALID_PARAMS",
			"Invalid competition parameters",
			"사용법: `!대회 <create|status|blackout|update>`")
		return
	}

	subCommand := params[0]
	switch subCommand {
	case "create":
		ch.handleCompetitionCreate(s, m, params[1:])
	case "status":
		ch.handleCompetitionStatus(s, m)
	case "blackout":
		ch.handleCompetitionBlackout(s, m, params[1:])
	case "update":
		ch.handleCompetitionUpdate(s, m, params[1:])
	default:
		err := errors.NewValidationError("COMPETITION_UNKNOWN_COMMAND",
			fmt.Sprintf("Unknown competition command: %s", subCommand),
			"알 수 없는 명령어입니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
	}
}

func (ch *CompetitionHandler) handleCompetitionCreate(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	if len(params) < 3 {
		err := errors.NewValidationError("COMPETITION_CREATE_INVALID_PARAMS",
			"Invalid competition create parameters",
			constants.MsgCompetitionCreateUsage)
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	name := params[0]
	startDateStr := params[1]
	endDateStr := params[2]

	startDate, endDate, err := utils.ValidateAndParseCompetitionDates(name, startDateStr, endDateStr)
	if err != nil {
		errorHandlers.Validation().HandleInvalidParams("INVALID_COMPETITION_DATES",
			fmt.Sprintf("Invalid competition dates: %v", err),
			fmt.Sprintf("날짜 오류: %v", err))
		return
	}

	err = ch.commandHandler.deps.Storage.CreateCompetition(name, startDate, endDate)
	if err != nil {
		errorHandlers.System().HandleCompetitionCreateFailed(err)
		return
	}

	// 봇 상태 업데이트
	ch.commandHandler.deps.UpdateBotStatus()

	// 대회 생성 텔레메트리 전송
	if ch.commandHandler.deps.MetricsClient != nil {
		participantCount := len(ch.commandHandler.deps.Storage.GetParticipants())
		ch.commandHandler.deps.MetricsClient.SendCompetitionMetric("created", participantCount)
	}

	blackoutStart := endDate.AddDate(0, 0, -constants.BlackoutDays)
	response := fmt.Sprintf(constants.MsgCompetitionCreateSuccess,
		name,
		utils.FormatDate(startDate),
		utils.FormatDate(endDate),
		utils.FormatDate(blackoutStart))

	errors.SendDiscordSuccess(s, m.ChannelID, response)
}

func (ch *CompetitionHandler) handleCompetitionStatus(s *discordgo.Session, m *discordgo.MessageCreate) {
	competition := ch.commandHandler.deps.Storage.GetCompetition()
	if competition == nil {
		err := errors.NewNotFoundError("NO_ACTIVE_COMPETITION",
			"No active competition found",
			"활성화된 대회가 없습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	now := utils.GetCurrentTimeKST()
	status := constants.StatusActive
	if now.Before(competition.StartDate) {
		status = constants.StatusInactive
	} else if now.After(competition.EndDate) {
		status = constants.StatusInactive
	}

	blackoutStatus := constants.StatusVisible
	if ch.commandHandler.deps.Storage.IsBlackoutPeriod() {
		blackoutStatus = constants.StatusHidden
	}

	response := fmt.Sprintf(constants.MsgCompetitionStatus,
		competition.Name,
		utils.FormatDate(competition.StartDate),
		utils.FormatDate(competition.EndDate),
		blackoutStatus,
		status,
		len(ch.commandHandler.deps.Storage.GetParticipants()))

	if _, err := s.ChannelMessageSend(m.ChannelID, response); err != nil {
		utils.Error("Failed to send competition status message: %v", err)
	}
}

func (ch *CompetitionHandler) handleCompetitionBlackout(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	if len(params) == 0 {
		err := errors.NewValidationError("BLACKOUT_INVALID_PARAMS",
			"Invalid blackout parameters",
			"사용법: `!대회 blackout <on|off>`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	setting := strings.ToLower(params[0])
	var visible bool

	switch setting {
	case "on":
		visible = false
	case "off":
		visible = true
	default:
		err := errors.NewValidationError("BLACKOUT_INVALID_SETTING",
			fmt.Sprintf("Invalid blackout setting: %s", setting),
			"올바른 설정값을 입력하세요: `on` 또는 `off`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	err := ch.commandHandler.deps.Storage.SetScoreboardVisibility(visible)
	if err != nil {
		botErr := errors.NewSystemError("BLACKOUT_SETTING_FAILED",
			"Failed to set scoreboard visibility", err)
		botErr.UserMsg = "설정 변경에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	status := "공개"
	if !visible {
		status = "비공개"
	}

	message := fmt.Sprintf("스코어보드가 **%s**로 설정되었습니다.", status)
	errors.SendDiscordSuccess(s, m.ChannelID, message)
}

func (ch *CompetitionHandler) handleCompetitionUpdate(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	if len(params) < 2 {
		err := errors.NewValidationError("COMPETITION_UPDATE_INVALID_PARAMS",
			"Invalid competition update parameters",
			"사용법: `!대회 update <필드> <값>`\n필드: name, start, end\n예시: `!대회 update name 대회명`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	field := strings.ToLower(params[0])
	value := strings.Join(params[1:], " ")

	competition := ch.commandHandler.deps.Storage.GetCompetition()
	if competition == nil {
		err := errors.NewNotFoundError("NO_ACTIVE_COMPETITION",
			"No active competition found",
			"수정할 대회가 없습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	switch field {
	case "name":
		ch.handleUpdateName(s, m, value, competition.Name)
	case "start":
		ch.handleUpdateStartDate(s, m, value, competition)
	case "end":
		ch.handleUpdateEndDate(s, m, value, competition)
	default:
		err := errors.NewValidationError("INVALID_UPDATE_FIELD",
			fmt.Sprintf("Invalid field: %s", field),
			"올바르지 않은 필드입니다. 사용 가능한 필드: name, start, end")
		errors.HandleDiscordError(s, m.ChannelID, err)
	}
}

func (ch *CompetitionHandler) handleUpdateName(s *discordgo.Session, m *discordgo.MessageCreate, newName, oldName string) {
	if newName == "" {
		err := errors.NewValidationError("EMPTY_COMPETITION_NAME",
			"Competition name cannot be empty",
			"대회명이 비어있습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	err := ch.commandHandler.deps.Storage.UpdateCompetitionName(newName)
	if err != nil {
		botErr := errors.NewSystemError("COMPETITION_UPDATE_FAILED",
			"Failed to update competition name", err)
		botErr.UserMsg = "대회명 수정에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	// 봇 상태 업데이트
	ch.commandHandler.deps.UpdateBotStatus()

	message := fmt.Sprintf(constants.MsgCompetitionUpdateSuccess, "대회명")
	errors.SendDiscordSuccess(s, m.ChannelID, message)
}

func (ch *CompetitionHandler) handleUpdateStartDate(s *discordgo.Session, m *discordgo.MessageCreate, dateStr string, competition *models.Competition) {
	ch.updateCompetitionDate(s, m, dateStr, competition, true)
}

func (ch *CompetitionHandler) handleUpdateEndDate(s *discordgo.Session, m *discordgo.MessageCreate, dateStr string, competition *models.Competition) {
	ch.updateCompetitionDate(s, m, dateStr, competition, false)
}

// updateCompetitionDate 대회 시작 날짜와 종료 날짜를 업데이트합니다
func (ch *CompetitionHandler) updateCompetitionDate(s *discordgo.Session, m *discordgo.MessageCreate, dateStr string, competition *models.Competition, isStartDate bool) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// Determine field name and labels based on whether it's start or end date
	var fieldName, fieldLabel, formatLabel, errorMsg string
	if isStartDate {
		fieldName = "start"
		fieldLabel = "START"
		formatLabel = "시작일"
		errorMsg = "시작일 수정에 실패했습니다."
	} else {
		fieldName = "end"
		fieldLabel = "END"
		formatLabel = "종료일"
		errorMsg = "종료일 수정에 실패했습니다."
	}

	// Parse and validate the date
	parsedDate, err := utils.ParseDateWithValidation(dateStr, fieldName)
	if err != nil {
		errorHandlers.Validation().HandleInvalidDateFormat(fieldLabel)
		return
	}

	// Validate date range
	var validRange bool
	if isStartDate {
		validRange = utils.IsValidDateRange(parsedDate, competition.EndDate)
	} else {
		validRange = utils.IsValidDateRange(competition.StartDate, parsedDate)
	}

	if !validRange {
		errorHandlers.Validation().HandleInvalidDateRange()
		return
	}

	// Update the date in storage
	if isStartDate {
		err = ch.commandHandler.deps.Storage.UpdateCompetitionStartDate(parsedDate)
	} else {
		err = ch.commandHandler.deps.Storage.UpdateCompetitionEndDate(parsedDate)
	}

	if err != nil {
		botErr := errors.NewSystemError("COMPETITION_UPDATE_FAILED",
			fmt.Sprintf("Failed to update competition %s date", fieldName), err)
		botErr.UserMsg = errorMsg
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	// Send success message
	message := fmt.Sprintf(constants.MsgCompetitionUpdateSuccess, formatLabel)
	errors.SendDiscordSuccess(s, m.ChannelID, message)
}
