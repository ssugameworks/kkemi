package bot

import (
	"github.com/ssugameworks/Discord-Bot/api"
	"github.com/ssugameworks/Discord-Bot/constants"
	"github.com/ssugameworks/Discord-Bot/errors"
	"github.com/ssugameworks/Discord-Bot/utils"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	deps               *CommandDependencies
	competitionHandler *CompetitionHandler
}

func NewCommandHandler(deps *CommandDependencies) *CommandHandler {
	ch := &CommandHandler{
		deps: deps,
	}
	ch.competitionHandler = NewCompetitionHandler(ch)
	return ch
}

// HandleMessage Discord 메시지를 처리합니다
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
		utils.Debug("DM received from %s", m.Author.Username)
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
	// 명령어 사용 텔레메트리 전송
	isAdmin := ch.isAdmin(s, m)
	if ch.deps.MetricsClient != nil {
		ch.deps.MetricsClient.SendCommandMetric(command, isAdmin)
	}

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
		if _, err := s.ChannelMessageSend(m.ChannelID, constants.MsgScoreboardDMOnly); err != nil {
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
		utils.Error("DISCORD API ERROR: Failed to send help message: %v", err)
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

	// 4. 숭실대학교 소속 검증
	organizationID, ok := ch.validateUniversityAffiliation(baekjoonID, errorHandlers)
	if !ok {
		return
	}

	// 5. 참가자 등록
	if !ch.registerParticipant(name, baekjoonID, userInfo, organizationID, errorHandlers) {
		return
	}

	// 6. 성공 메시지 전송
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
	competition := ch.deps.Storage.GetCompetition()
	if competition == nil {
		errorHandlers.Data().HandleNoActiveCompetition()
		return false
	}

	now := utils.GetCurrentTimeKST()
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
	info, err := ch.deps.APIClient.GetUserInfo(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return nil, false
	}

	// solved.ac 추가 정보 조회 (본명 확인용)
	additionalInfo, err := ch.deps.APIClient.GetUserAdditionalInfo(baekjoonID)
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

	// 스프레드시트에서 이름 검증
	if ch.deps.SheetsClient != nil {
		isInList, err := ch.deps.SheetsClient.IsNameInParticipantList(name)
		if err != nil {
			utils.Warn("Failed to check participant list: %v", err)
			errorHandlers.System().HandleSystemError("SHEETS_CHECK_FAILED",
				"Failed to verify participant eligibility",
				constants.ErrorSheetsCheckFailed, err)
			return nil, false
		}
		if !isInList {
			// 스프레드시트에 없으면 백업 명단에서 확인
			if utils.IsNameInBackupList(name) {
				utils.Info("Name '%s' found in backup participant list", name)
			} else {
				errorHandlers.Validation().HandleInvalidParams("NAME_NOT_IN_LIST",
					"Name not found in participant list",
					fmt.Sprintf(constants.ErrorNameNotInList, name))
				return nil, false
			}
		}
		utils.Info("Name '%s' verified in participant list", name)
	} else {
		// SheetsClient가 없으면 백업 명단에서만 확인
		utils.Warn("SheetsClient not available, using backup participant list")
		if utils.IsNameInBackupList(name) {
			utils.Info("Name '%s' found in backup participant list", name)
		} else {
			errorHandlers.Validation().HandleInvalidParams("NAME_NOT_IN_LIST",
				"Name not found in participant list",
				fmt.Sprintf(constants.ErrorNameNotInList, name))
			return nil, false
		}
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

// validateUniversityAffiliation 사용자의 학교 소속을 검증합니다
func (ch *CommandHandler) validateUniversityAffiliation(baekjoonID string, errorHandlers *utils.ErrorHandlerFactory) (organizationID int, ok bool) {
	// solved.ac에서 사용자의 조직 정보 조회
	organizations, err := ch.deps.APIClient.GetUserOrganizations(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return 0, false
	}

	// 특정 학교 소속인지 확인
	for _, org := range organizations {
		if org.OrganizationID == constants.UniversityID {
			return constants.UniversityID, true
		}
	}

	// 특정 학교 소속이 아닌 경우 에러 메시지 전송
	errorHandlers.Validation().HandleInvalidParams("NOT_SOONGSIL_UNIVERSITY",
		"User is not affiliated with Soongsil University",
		constants.MsgRegisterNotSoongsilStudent)
	return 0, false
}

// registerParticipant 참가자를 등록합니다
func (ch *CommandHandler) registerParticipant(name, baekjoonID string, userInfo interface{}, organizationID int, errorHandlers *utils.ErrorHandlerFactory) bool {
	// Type assertion to get the actual type
	info, ok := userInfo.(*api.UserInfo)
	if !ok {
		errorHandlers.System().HandleSystemError("TYPE_ASSERTION_FAILED", "Failed to process user info", "내부 처리 오류가 발생했습니다.", nil)
		return false
	}

	err := ch.deps.Storage.AddParticipant(name, baekjoonID, info.Tier, info.Rating, organizationID)
	if err != nil {
		errorHandlers.Data().HandleParticipantAlreadyExists(baekjoonID)
		return false
	}

	// 참가자 등록 텔레메트리 전송
	if ch.deps.MetricsClient != nil {
		participantCount := len(ch.deps.Storage.GetParticipants())
		ch.deps.MetricsClient.SendCompetitionMetric("participant_registered", participantCount)
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

	tierName := ch.deps.TierManager.GetTierName(info.Tier)
	colorCode := ch.deps.TierManager.GetTierANSIColor(info.Tier)

	// 사용자 리그 결정 및 이름 가져오기
	userLeague := ch.deps.ScoreCalculator.GetUserLeague(info.Tier)
	leagueName := ch.deps.ScoreCalculator.GetLeagueName(userLeague)

	response := fmt.Sprintf("```ansi\n"+constants.MsgRegisterSuccess+"\n```",
		colorCode, name, tierName, ch.deps.TierManager.GetANSIReset(), leagueName)

	if _, err := s.ChannelMessageSend(channelID, response); err != nil {
		utils.Error("DISCORD API ERROR: Failed to send registration response: %v", err)
	}
}

func (ch *CommandHandler) handleScoreboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	utils.Info("Scoreboard command received from user: %s (ID: %s)", m.Author.Username, m.Author.ID)
	utils.Info("Guild ID: %s, Channel ID: %s", m.GuildID, m.ChannelID)

	// 관리자 권한 확인
	isAdmin := ch.isAdmin(s, m)
	utils.Info("User %s admin status: %t", m.Author.Username, isAdmin)

	if !isAdmin {
		utils.Warn("User %s attempted to use scoreboard without admin permissions", m.Author.Username)
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	// 스코어보드 생성 성능 측정 시작
	startTime := time.Now()
	embed, err := ch.deps.ScoreboardManager.GenerateScoreboard(isAdmin)
	duration := time.Since(startTime)

	// 스코어보드 성능 텔레메트리 전송
	if ch.deps.MetricsClient != nil {
		ch.deps.MetricsClient.SendPerformanceMetric("scoreboard_generation", duration, err == nil)
	}

	if err != nil {
		utils.Error("Failed to generate scoreboard: %v", err)
		errorHandlers.System().HandleScoreboardGenerationFailed(err)
		return
	}

	utils.Info("Scoreboard generated successfully, sending to channel %s", m.ChannelID)

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, embed); err != nil {
		utils.Error("DISCORD API ERROR: Failed to send scoreboard embed: %v", err)
	} else {
		utils.Info("Scoreboard sent successfully")
	}
}

func (ch *CommandHandler) handleParticipants(s *discordgo.Session, m *discordgo.MessageCreate) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)

	// 관리자 권한 확인
	if !ch.isAdmin(s, m) {
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	participants := ch.deps.Storage.GetParticipants()
	if len(participants) == 0 {
		errors.SendDiscordInfo(s, m.ChannelID, constants.MsgParticipantsEmpty)
		return
	}

	var sb strings.Builder
	sb.WriteString("```ansi\n")

	for i, p := range participants {
		tierName := ch.deps.TierManager.GetTierName(p.StartTier)
		colorCode := ch.deps.TierManager.GetTierANSIColor(p.StartTier)
		sb.WriteString(fmt.Sprintf("%s%d. %s - %s%s\n",
			colorCode, i+1, p.BaekjoonID, tierName, ch.deps.TierManager.GetANSIReset()))
	}

	sb.WriteString("```")
	if _, err := s.ChannelMessageSend(m.ChannelID, sb.String()); err != nil {
		utils.Error("DISCORD API ERROR: Failed to send participants list: %v", err)
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
	err := ch.deps.Storage.RemoveParticipant(baekjoonID)
	if err != nil {
		errorHandlers.Data().HandleParticipantNotFound(baekjoonID)
		return
	}

	response := fmt.Sprintf(constants.MsgRemoveSuccess, baekjoonID)
	if err := errors.SendDiscordSuccess(s, m.ChannelID, response); err != nil {
		utils.Error("Failed to send participant removal response: %v", err)
	}
}

// isAdmin 사용자가 서버 관리자 권한을 가지고 있는지 확인합니다
func (ch *CommandHandler) isAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {

	// DM에서는 관리자 권한 없음
	if m.GuildID == "" {
		utils.Info("User is in DM, no admin permissions")
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
		utils.Info("User %s is the guild owner - granting admin access", m.Author.Username)
		return true
	}

	// 멤버 정보 가져오기
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil || member == nil {
		utils.Warn("Cannot get member information for %s: %v", m.Author.Username, err)
		return false
	}

	utils.Info("Member found with %d roles", len(member.Roles))

	// 멤버의 역할들을 확인
	for _, roleID := range member.Roles {
		role, err := s.State.Role(m.GuildID, roleID)
		if err != nil {
			utils.Warn("Cannot get role %s: %v", roleID, err)
			continue
		}

		utils.Info("Checking role: %s (ID: %s), Permissions: %d", role.Name, role.ID, role.Permissions)

		// 관리자 권한(ADMINISTRATOR) 확인
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			utils.Info("User %s has ADMINISTRATOR permission through role %s - granting admin access", m.Author.Username, role.Name)
			return true
		}
	}

	utils.Info("User %s has no admin permissions", m.Author.Username)
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

	if cachedClient, ok := ch.deps.APIClient.(*api.CachedSolvedACClient); ok {
		stats := cachedClient.GetCacheStats()

		message := fmt.Sprintf("```\n📊 API Cache Statistics\n\n"+
			"Total API Calls: %d\n"+
			"Cache Hits: %d\n"+
			"Cache Misses: %d\n"+
			"Hit Rate: %.2f%%\n\n"+
			"Cached Items:\n"+
			"  - User Info: %d\n"+
			"  - User Top100: %d\n"+
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
