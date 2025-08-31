package bot

import (
	"discord-bot/api"
	"discord-bot/constants"
	"discord-bot/errors"
	"discord-bot/interfaces"
	"discord-bot/models"
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
	tierManager        *models.TierManager
}

func NewCommandHandler(storage interfaces.StorageRepository, apiClient interfaces.APIClient, scoreboardManager *ScoreboardManager, tierManager *models.TierManager) *CommandHandler {
	ch := &CommandHandler{
		storage:           storage,
		scoreboardManager: scoreboardManager,
		client:            apiClient,
		tierManager:       tierManager,
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
	case "cache", "캐시":
		ch.handleCacheStats(s, m)
	case "ping":
		ch.handlePing(s, m)
	}
}

// handleScoreboardCommand 스코어보드 명령어를 처리합니다 (DM 체크 포함)
func (ch *CommandHandler) handleScoreboardCommand(s *discordgo.Session, m *discordgo.MessageCreate, isDM bool) {
	if isDM {
		if err := errors.SendDiscordInfo(s, m.ChannelID, constants.MsgScoreboardDMOnly); err != nil {
			utils.Error("Failed to send DM response: %v", err)
		}
		return
	}
	ch.handleScoreboard(s, m)
}

// handlePing ping 명령어를 처리합니다
func (ch *CommandHandler) handlePing(s *discordgo.Session, m *discordgo.MessageCreate) {
	if err := errors.SendDiscordInfo(s, m.ChannelID, constants.MsgPong); err != nil {
		utils.Error("Failed to send ping response: %v", err)
	}
}

func (ch *CommandHandler) handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	if _, err := s.ChannelMessageSend(m.ChannelID, constants.HelpMessage); err != nil {
		fmt.Printf("DISCORD API ERROR: Failed to send help message: %v\n", err)
		utils.Error("Failed to send help message: %v", err)
	}
}

func (ch *CommandHandler) handleRegister(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// 1. 기본 매개변수 검증
	name, baekjoonID, ok := ch.validateRegisterParams(params, errorHandlers)
	if !ok {
		return
	}

	// 2. 대회 상태 확인
	if !ch.validateCompetitionStatus(errorHandlers) {
		return
	}

	// 3. solved.ac 사용자 정보 조회 및 검증
	userInfo, ok := ch.validateSolvedACUser(name, baekjoonID, errorHandlers)
	if !ok {
		return
	}

	// 4. 참가자 등록
	if !ch.registerParticipant(name, baekjoonID, userInfo, errorHandlers) {
		return
	}

	// 5. 성공 메시지 전송
	ch.sendRegistrationSuccess(s, m.ChannelID, name, userInfo)
}

// validateRegisterParams 등록 매개변수를 검증합니다
func (ch *CommandHandler) validateRegisterParams(params []string, errorHandlers *utils.ErrorHandlerFactory) (name, baekjoonID string, ok bool) {
	if len(params) < 2 {
		errorHandlers.Validation().HandleInvalidParams("REGISTER_INVALID_PARAMS",
			"Invalid register parameters",
			constants.MsgRegisterUsage)
		return "", "", false
	}
	return params[0], params[1], true
}

// validateCompetitionStatus 대회 상태를 확인합니다
func (ch *CommandHandler) validateCompetitionStatus(errorHandlers *utils.ErrorHandlerFactory) bool {
	competition := ch.storage.GetCompetition()
	if competition == nil {
		errorHandlers.Data().HandleNoActiveCompetition()
		return false
	}

	now := time.Now()
	if now.Before(competition.StartDate) {
		errorHandlers.Validation().HandleInvalidParams("REGISTRATION_NOT_STARTED",
			"Registration not available before competition starts",
			fmt.Sprintf(constants.MsgRegisterNotStarted, 
				utils.FormatDateTime(competition.StartDate)))
		return false
	}
	return true
}

// validateSolvedACUser solved.ac 사용자 정보를 조회하고 이름을 검증합니다
func (ch *CommandHandler) validateSolvedACUser(name, baekjoonID string, errorHandlers *utils.ErrorHandlerFactory) (userInfo interface{}, ok bool) {
	// solved.ac 사용자 정보 조회
	info, err := ch.client.GetUserInfo(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return nil, false
	}

	// solved.ac 추가 정보 조회 (본명 확인용)
	additionalInfo, err := ch.client.GetUserAdditionalInfo(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return nil, false
	}

	// solved.ac에 등록된 이름 추출 및 검증
	solvedacName := ch.extractSolvedACName(additionalInfo, errorHandlers)
	if solvedacName == "" {
		return nil, false
	}

	// 입력한 이름과 solved.ac 이름 일치 확인
	if name != solvedacName {
		errorHandlers.Validation().HandleInvalidParams("NAME_MISMATCH",
			"Name does not match solved.ac profile",
			fmt.Sprintf(constants.MsgRegisterNameMismatch, name, solvedacName))
		return nil, false
	}

	return info, true
}

// extractSolvedACName solved.ac 추가 정보에서 이름을 추출합니다
func (ch *CommandHandler) extractSolvedACName(additionalInfo interface{}, errorHandlers *utils.ErrorHandlerFactory) string {
	// Type assertion to get the actual type
	info, ok := additionalInfo.(*api.UserAdditionalInfo)
	if !ok {
		errorHandlers.System().HandleSystemError("TYPE_ASSERTION_FAILED", "Failed to process user additional info", "내부 처리 오류가 발생했습니다.", nil)
		return ""
	}

	if info.NameNative != nil && *info.NameNative != "" {
		return *info.NameNative
	} else if info.Name != nil && *info.Name != "" {
		return *info.Name
	} else {
		errorHandlers.Validation().HandleInvalidParams("NO_SOLVEDAC_NAME",
			"No name registered in solved.ac",
			constants.MsgRegisterNoSolvedacName)
		return ""
	}
}

// registerParticipant 참가자를 등록합니다
func (ch *CommandHandler) registerParticipant(name, baekjoonID string, userInfo interface{}, errorHandlers *utils.ErrorHandlerFactory) bool {
	// Type assertion to get the actual type
	info, ok := userInfo.(*api.UserInfo)
	if !ok {
		errorHandlers.System().HandleSystemError("TYPE_ASSERTION_FAILED", "Failed to process user info", "내부 처리 오류가 발생했습니다.", nil)
		return false
	}

	err := ch.storage.AddParticipant(name, baekjoonID, info.Tier, info.Rating)
	if err != nil {
		errorHandlers.Data().HandleParticipantAlreadyExists(baekjoonID)
		return false
	}
	return true
}

// sendRegistrationSuccess 등록 성공 메시지를 전송합니다
func (ch *CommandHandler) sendRegistrationSuccess(s *discordgo.Session, channelID, name string, userInfo interface{}) {
	// Type assertion to get the actual type
	info, ok := userInfo.(*api.UserInfo)
	if !ok {
		utils.Error("Failed to send registration success: type assertion failed")
		return
	}

	tierName := ch.tierManager.GetTierName(info.Tier)
	colorCode := ch.tierManager.GetTierANSIColor(info.Tier)

	response := fmt.Sprintf("```ansi\n"+constants.MsgRegisterSuccess+"\n```",
		colorCode, name, tierName, ch.tierManager.GetANSIReset())

	if _, err := s.ChannelMessageSend(channelID, response); err != nil {
		fmt.Printf("DISCORD API ERROR: Failed to send registration response: %v\n", err)
		utils.Error("Failed to send registration response: %v", err)
	}
}

func (ch *CommandHandler) handleScoreboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// 관리자 권한 확인
	if !ch.isAdmin(s, m) {
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	isAdmin := ch.isAdmin(s, m)
	embed, err := ch.scoreboardManager.GenerateScoreboard(isAdmin)
	if err != nil {
		errorHandlers.System().HandleScoreboardGenerationFailed(err)
		return
	}

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, embed); err != nil {
		fmt.Printf("DISCORD API ERROR: Failed to send scoreboard embed: %v\n", err)
		utils.Error("Failed to send scoreboard embed message: %v", err)
	}
}

func (ch *CommandHandler) handleParticipants(s *discordgo.Session, m *discordgo.MessageCreate) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// 관리자 권한 확인
	if !ch.isAdmin(s, m) {
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	participants := ch.storage.GetParticipants()
	if len(participants) == 0 {
		errors.SendDiscordInfo(s, m.ChannelID, constants.MsgParticipantsEmpty)
		return
	}

	var sb strings.Builder
	sb.WriteString("```ansi\n")

	for i, p := range participants {
		tierName := ch.tierManager.GetTierName(p.StartTier)
		colorCode := ch.tierManager.GetTierANSIColor(p.StartTier)
		sb.WriteString(fmt.Sprintf("%s%d. %s - %s%s\n",
			colorCode, i+1, p.BaekjoonID, tierName, ch.tierManager.GetANSIReset()))
	}

	sb.WriteString("```")
	if _, err := s.ChannelMessageSend(m.ChannelID, sb.String()); err != nil {
		fmt.Printf("DISCORD API ERROR: Failed to send participants list: %v\n", err)
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
			constants.MsgRemoveUsage)
		return
	}

	baekjoonID := params[0]

	// 백준 ID 유효성 검사
	if !utils.IsValidBaekjoonID(baekjoonID) {
		errorHandlers.Validation().HandleInvalidParams("REMOVE_INVALID_BAEKJOON_ID",
			"Invalid Baekjoon ID format",
			constants.MsgRemoveInvalidBaekjoonID)
		return
	}

	// 참가자 삭제
	err := ch.storage.RemoveParticipant(baekjoonID)
	if err != nil {
		errorHandlers.Data().HandleParticipantNotFound(baekjoonID)
		return
	}

	response := fmt.Sprintf(constants.MsgRemoveSuccess, baekjoonID)
	if err := errors.SendDiscordSuccess(s, m.ChannelID, response); err != nil {
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

// handleCacheStats 캐시 통계를 조회합니다
func (ch *CommandHandler) handleCacheStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// 관리자 권한 확인
	if !ch.isAdmin(s, m) {
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	if cachedClient, ok := ch.client.(*api.CachedSolvedACClient); ok {
		stats := cachedClient.GetCacheStats()
		
		message := fmt.Sprintf("```\n📊 API Cache Statistics\n\n" +
			"Total API Calls: %d\n" +
			"Cache Hits: %d\n" +
			"Cache Misses: %d\n" +
			"Hit Rate: %.2f%%\n\n" +
			"Cached Items:\n" +
			"  - User Info: %d\n" +
			"  - User Top100: %d\n" +
			"  - User Additional: %d\n```",
			stats.TotalCalls, stats.CacheHits, stats.CacheMisses, stats.HitRate,
			stats.UserInfoCached, stats.UserTop100Cached, stats.UserAdditionalCached)
		
		if err := errors.SendDiscordInfo(s, m.ChannelID, message); err != nil {
			utils.Error("Failed to send cache stats response: %v", err)
		}
	} else {
		if err := errors.SendDiscordWarning(s, m.ChannelID, "캐시가 비활성화되어 있습니다."); err != nil {
			utils.Error("Failed to send cache disabled warning: %v", err)
		}
	}
}



