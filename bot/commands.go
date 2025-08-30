package bot

import (
	"discord-bot/constants"
	"discord-bot/errors"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/scoring"
	"discord-bot/utils"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	storage            interfaces.StorageRepository
	scoreboardManager  *ScoreboardManager
	client             interfaces.APIClient
	competitionHandler *CompetitionHandler
}

func NewCommandHandler(storage interfaces.StorageRepository, apiClient interfaces.APIClient, scoreboardManager *ScoreboardManager) *CommandHandler {
	ch := &CommandHandler{
		storage:           storage,
		scoreboardManager: scoreboardManager,
		client:            apiClient,
	}
	ch.competitionHandler = NewCompetitionHandler(ch)
	return ch
}

func (ch *CommandHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if ch.shouldIgnoreMessage(s, m) {
		return
	}

	command, params, isDM := ch.parseMessage(m)
	if command == "" {
		return
	}

	ch.routeCommand(s, m, command, params, isDM)
}

// shouldIgnoreMessage 메시지를 무시해야 하는지 확인합니다
func (ch *CommandHandler) shouldIgnoreMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// 봇 자신의 메시지는 무시
	if m.Author.ID == s.State.User.ID {
		return true
	}

	// DM 디버깅 로그
	if m.GuildID == "" {
		fmt.Printf(constants.DMReceivedTemplate, m.Content, m.Author.Username)
	}

	return false
}

// parseMessage 메시지를 파싱하여 명령어와 매개변수를 추출합니다
func (ch *CommandHandler) parseMessage(m *discordgo.MessageCreate) (command string, params []string, isDM bool) {
	content := strings.TrimSpace(m.Content)
	if !strings.HasPrefix(content, constants.CommandPrefix) {
		return "", nil, false
	}

	args := strings.Fields(content)
	if len(args) == 0 {
		return "", nil, false
	}

	command = args[0][constants.CommandPrefixLength:]
	params = args[1:]
	isDM = m.GuildID == ""

	return command, params, isDM
}

// routeCommand 명령어를 해당 핸들러로 라우팅합니다
func (ch *CommandHandler) routeCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, params []string, isDM bool) {
	switch command {
	case "help", "도움말":
		ch.handleHelp(s, m)
	case "register", "등록":
		ch.handleRegister(s, m, params)
	case "scoreboard", "스코어보드":
		ch.handleScoreboardCommand(s, m, isDM)
	case "competition", "대회":
		ch.competitionHandler.HandleCompetition(s, m, params)
	case "participants", "참가자":
		ch.handleParticipants(s, m)
	case "remove", "삭제":
		ch.handleRemoveParticipant(s, m, params)
	case "test":
		ch.handleTest(s, m, params)
	case "ping":
		ch.handlePing(s, m)
	}
}

// handleScoreboardCommand 스코어보드 명령어를 처리합니다 (DM 체크 포함)
func (ch *CommandHandler) handleScoreboardCommand(s *discordgo.Session, m *discordgo.MessageCreate, isDM bool) {
	if isDM {
		if _, err := s.ChannelMessageSend(m.ChannelID, "❌ 스코어보드는 서버에서만 확인할 수 있습니다."); err != nil {
			utils.Error("Failed to send DM response: %v", err)
		}
		return
	}
	ch.handleScoreboard(s, m)
}

// handlePing ping 명령어를 처리합니다
func (ch *CommandHandler) handlePing(s *discordgo.Session, m *discordgo.MessageCreate) {
	if _, err := s.ChannelMessageSend(m.ChannelID, "Pong! 🏓"); err != nil {
		utils.Error("Failed to send ping response: %v", err)
	}
}

func (ch *CommandHandler) handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	helpText := `🤖 **알고리즘 경진대회 봇 명령어**

**참가자 명령어:**
• ` + "`!등록 <이름> <백준ID>`" + ` - 대회 등록 신청 (대회 시작 후, solved.ac 등록 이름과 일치해야 함)
• ` + "`!스코어보드`" + ` - 현재 스코어보드 확인
• ` + "`!참가자`" + ` - 참가자 목록 확인

**관리자 명령어:**
• ` + "`!대회 create <대회명> <시작일> <종료일>`" + ` - 대회 생성 (YYYY-MM-DD 형식)
• ` + "`!대회 status`" + ` - 대회 상태 확인
• ` + "`!대회 blackout <on/off>`" + ` - 스코어보드 공개/비공개 설정
• ` + "`!대회 update <필드> <값>`" + ` - 대회 정보 수정 (name, start, end)
• ` + "`!삭제 <백준ID>`" + ` - 참가자 삭제

**기타:**
• ` + "`!test <백준ID>`" + ` - 사용자 본명 확인
• ` + "`!ping`" + ` - 봇 응답 확인
• ` + "`!도움말`" + ` - 도움말 표시`

	if _, err := s.ChannelMessageSend(m.ChannelID, helpText); err != nil {
		utils.Error("Failed to send help message: %v", err)
	}
}

func (ch *CommandHandler) handleRegister(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	if len(params) < 2 {
		errorHandlers.Validation().HandleInvalidParams("REGISTER_INVALID_PARAMS",
			"Invalid register parameters",
			"사용법: `!등록 <이름> <백준ID>`")
		return
	}

	name := params[0]
	baekjoonID := params[1]

	// 대회가 존재하고 시작되었는지 확인
	competition := ch.storage.GetCompetition()
	if competition == nil {
		errorHandlers.Data().HandleNoActiveCompetition()
		return
	}

	now := time.Now()
	if now.Before(competition.StartDate) {
		errorHandlers.Validation().HandleInvalidParams("REGISTRATION_NOT_STARTED",
			"Registration not available before competition starts",
			fmt.Sprintf("대회가 아직 시작되지 않았습니다. 등록은 %s부터 가능합니다.", 
				utils.FormatDateTime(competition.StartDate)))
		return
	}

	// solved.ac 사용자 정보 조회
	userInfo, err := ch.client.GetUserInfo(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return
	}

	// solved.ac 추가 정보 조회 (본명 확인용)
	additionalInfo, err := ch.client.GetUserAdditionalInfo(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return
	}

	// solved.ac에 등록된 이름과 비교
	var solvedacName string
	if additionalInfo.NameNative != nil && *additionalInfo.NameNative != "" {
		solvedacName = *additionalInfo.NameNative
	} else if additionalInfo.Name != nil && *additionalInfo.Name != "" {
		solvedacName = *additionalInfo.Name
	} else {
		errorHandlers.Validation().HandleInvalidParams("NO_SOLVEDAC_NAME",
			"No name registered in solved.ac",
			"solved.ac에 이름이 등록되지 않았습니다. solved.ac 프로필에서 본명 또는 영문 이름을 등록한 후 다시 시도해주세요.")
		return
	}

	// 입력한 이름과 solved.ac 이름이 일치하는지 확인
	if name != solvedacName {
		errorHandlers.Validation().HandleInvalidParams("NAME_MISMATCH",
			"Name does not match solved.ac profile",
			fmt.Sprintf("입력한 이름 '%s'이 solved.ac에 등록된 이름 '%s'과 일치하지 않습니다.", 
				name, solvedacName))
		return
	}

	err = ch.storage.AddParticipant(name, baekjoonID, userInfo.Tier, userInfo.Rating)
	if err != nil {
		errorHandlers.Data().HandleParticipantAlreadyExists(baekjoonID)
		return
	}

	tierName := getTierName(userInfo.Tier)
	tm := models.NewTierManager()
	colorCode := tm.GetTierANSIColor(userInfo.Tier)

	response := fmt.Sprintf("```ansi\n%s%s(%s)%s님 성공적으로 등록되었습니다!\n```",
		colorCode, name, tierName, tm.GetANSIReset())

	if _, err := s.ChannelMessageSend(m.ChannelID, response); err != nil {
		utils.Error("Failed to send registration response: %v", err)
	}
}

func (ch *CommandHandler) handleScoreboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	isAdmin := ch.isAdmin(s, m)
	embed, err := ch.scoreboardManager.GenerateScoreboard(isAdmin)
	if err != nil {
		errorHandlers.System().HandleScoreboardGenerationFailed(err)
		return
	}

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, embed); err != nil {
		utils.Error("Failed to send scoreboard embed message: %v", err)
	}
}

func (ch *CommandHandler) handleParticipants(s *discordgo.Session, m *discordgo.MessageCreate) {
	participants := ch.storage.GetParticipants()
	if len(participants) == 0 {
		errors.SendDiscordInfo(s, m.ChannelID, "참가자가 없습니다.")
		return
	}

	var sb strings.Builder
	sb.WriteString("```ansi\n")

	tm := models.NewTierManager()
	for i, p := range participants {
		tierName := getTierName(p.StartTier)
		colorCode := tm.GetTierANSIColor(p.StartTier)
		sb.WriteString(fmt.Sprintf("%s%d. %s (%s) - %s%s\n",
			colorCode, i+1, p.Name, p.BaekjoonID, tierName, tm.GetANSIReset()))
	}

	sb.WriteString("```")
	if _, err := s.ChannelMessageSend(m.ChannelID, sb.String()); err != nil {
		utils.Error("Failed to send participants list message: %v", err)
	}
}

func (ch *CommandHandler) handleRemoveParticipant(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// 관리자 권한 확인
	if !ch.isAdmin(s, m) {
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	// 파라미터 확인
	if len(params) < 1 {
		errorHandlers.Validation().HandleInvalidParams("REMOVE_INVALID_PARAMS",
			"Invalid remove parameters",
			"사용법: `!삭제 <백준ID>`")
		return
	}

	baekjoonID := params[0]

	// 백준 ID 유효성 검사
	if !utils.IsValidBaekjoonID(baekjoonID) {
		errorHandlers.Validation().HandleInvalidParams("REMOVE_INVALID_BAEKJOON_ID",
			"Invalid Baekjoon ID format",
			"유효하지 않은 백준 ID 형식입니다.")
		return
	}

	// 참가자 삭제
	err := ch.storage.RemoveParticipant(baekjoonID)
	if err != nil {
		errorHandlers.Data().HandleParticipantNotFound(baekjoonID)
		return
	}

	response := fmt.Sprintf("✅ **참가자 삭제 완료**\n🎯 백준ID: %s", baekjoonID)
	if _, err := s.ChannelMessageSend(m.ChannelID, response); err != nil {
		utils.Error("Failed to send participant removal response: %v", err)
	}
}

// isAdmin는 사용자가 서버 관리자 권한을 가지고 있는지 확인합니다
func (ch *CommandHandler) isAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// DM에서는 관리자 권한 없음
	if m.GuildID == "" {
		return false
	}

	// 길드 정보 가져오기
	guild, err := s.State.Guild(m.GuildID)
	if err != nil || guild == nil {
		utils.Warn("Cannot get guild information: %v", err)
		return false
	}

	// 서버 소유자인지 확인
	if m.Author.ID == guild.OwnerID {
		return true
	}

	// 멤버 정보 가져오기
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil || member == nil {
		utils.Warn("Cannot get member information: %v", err)
		return false
	}

	// 멤버의 역할들을 확인
	for _, roleID := range member.Roles {
		role, err := s.State.Role(m.GuildID, roleID)
		if err != nil {
			continue
		}

		// 관리자 권한(ADMINISTRATOR) 확인
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}

	return false
}

func (ch *CommandHandler) handleTest(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	if len(params) < 1 {
		errorHandlers.Validation().HandleInvalidParams("TEST_INVALID_PARAMS",
			"Invalid test parameters",
			"사용법: `!test <백준ID>`")
		return
	}

	baekjoonID := params[0]

	additionalInfo, err := ch.client.GetUserAdditionalInfo(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return
	}

	// 디버깅 로그 추가
	var nameNativeStr string
	if additionalInfo.NameNative != nil {
		nameNativeStr = *additionalInfo.NameNative
	}
	utils.Debug("Additional info result - NameNative: '%s'", nameNativeStr)

	var response string
	if additionalInfo.NameNative != nil && *additionalInfo.NameNative != "" {
		response = fmt.Sprintf("🔍 **사용자 정보**\n"+
			"📝 백준ID: `%s`\n"+
			"🏷️ 본명: `%s`",
			baekjoonID, *additionalInfo.NameNative)
	} else {
		response = fmt.Sprintf("🔍 **사용자 정보**\n"+
			"📝 백준ID: `%s`\n"+
			"🏷️ 본명: `등록되지 않음`",
			baekjoonID)
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, response); err != nil {
		utils.Error("Failed to send test response: %v", err)
	}
}


func getTierName(tier int) string {
	return scoring.GetTierName(tier)
}
