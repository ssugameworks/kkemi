package bot

import (
	"discord-bot/constants"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/utils"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/bwmarrin/discordgo"
)

type ScoreboardManager struct {
	storage     interfaces.StorageRepository
	calculator  interfaces.ScoreCalculator
	client      interfaces.APIClient
	tierManager *models.TierManager
}

func NewScoreboardManager(storage interfaces.StorageRepository, calculator interfaces.ScoreCalculator, client interfaces.APIClient, tierManager *models.TierManager) *ScoreboardManager {
	return &ScoreboardManager{
		storage:     storage,
		calculator:  calculator,
		client:      client,
		tierManager: tierManager,
	}
}

func (sm *ScoreboardManager) GetStorage() interfaces.StorageRepository {
	return sm.storage
}

func (sm *ScoreboardManager) GenerateScoreboard(isAdmin bool) (*discordgo.MessageEmbed, error) {
	competition := sm.storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		return nil, fmt.Errorf("활성화된 대회가 없습니다")
	}

	// 블랙아웃 체크
	if embed := sm.checkBlackoutPeriod(competition, isAdmin); embed != nil {
		return embed, nil
	}

	// 참가자 체크
	participants := sm.storage.GetParticipants()
	if embed := sm.checkEmptyParticipants(competition, participants); embed != nil {
		return embed, nil
	}

	// 점수 데이터 수집
	scores, err := sm.collectScoreData(participants)
	if err != nil {
		return nil, err
	}

	// 정렬 및 포맷팅
	sm.sortScores(scores)
	return sm.formatScoreboard(competition, scores, isAdmin), nil
}

// checkBlackoutPeriod 블랙아웃 기간인지 확인하고 해당 embed 반환
func (sm *ScoreboardManager) checkBlackoutPeriod(competition *models.Competition, isAdmin bool) *discordgo.MessageEmbed {
	if sm.storage.IsBlackoutPeriod() && competition.ShowScoreboard && !isAdmin {
		return &discordgo.MessageEmbed{
			Title:       constants.MsgScoreboardBlackout,
			Description: constants.MsgScoreboardBlackoutDesc,
			Color:       sm.tierManager.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// checkEmptyParticipants 참가자가 없는지 확인하고 해당 embed 반환
func (sm *ScoreboardManager) checkEmptyParticipants(competition *models.Competition, participants []models.Participant) *discordgo.MessageEmbed {
	if len(participants) == 0 {
		return &discordgo.MessageEmbed{
			Title:       fmt.Sprintf(constants.MsgScoreboardTitle, competition.Name),
			Description: constants.MsgScoreboardNoParticipants,
			Color:       sm.tierManager.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// collectScoreData 참가자들의 점수 데이터를 병렬로 수집합니다
func (sm *ScoreboardManager) collectScoreData(participants []models.Participant) ([]models.ScoreData, error) {
	if len(participants) == 0 {
		return []models.ScoreData{}, nil
	}

	// 메모리 사용량 최적화: 사전 할당된 슬라이스 사용
	scores := make([]models.ScoreData, 0, len(participants))
	scoreChan := make(chan models.ScoreData, len(participants))
	semaphore := make(chan struct{}, constants.MaxConcurrentRequests)
	var wg sync.WaitGroup
	var errorCount int64

	// 각 참가자에 대해 병렬로 점수 계산
	for _, participant := range participants {
		wg.Add(1)
		go func(p models.Participant) {
			defer wg.Done()

			// 동시 요청 수 제한
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			scoreData, err := sm.calculateParticipantScore(p)
			if err != nil {
				utils.Warn("Failed to calculate score for participant %s: %v", p.Name, err)
				atomic.AddInt64(&errorCount, 1)
				return
			}
			scoreChan <- scoreData
		}(participant)
	}

	// 고루틴들이 완료될 때까지 대기
	wg.Wait()
	close(scoreChan)

	// 결과 수집 - 메모리 효율적으로 수집
	for score := range scoreChan {
		scores = append(scores, score)
	}

	if errorCount > 0 {
		utils.Warn("Failed to calculate scores for %d participants", errorCount)
	}

	utils.Info("Successfully calculated scores for %d out of %d participants", len(scores), len(participants))
	return scores, nil
}

// calculateParticipantScore 개별 참가자의 점수를 계산합니다
func (sm *ScoreboardManager) calculateParticipantScore(participant models.Participant) (models.ScoreData, error) {
	userInfo, err := sm.client.GetUserInfo(participant.BaekjoonID)
	if err != nil {
		return models.ScoreData{}, err
	}

	top100, err := sm.client.GetUserTop100(participant.BaekjoonID)
	if err != nil {
		return models.ScoreData{}, err
	}

	score := sm.calculator.CalculateScoreWithTop100(top100, participant.StartTier, participant.StartProblemIDs)

	// 새로 푼 문제 수 계산 (현재 - 시작시점)
	newProblemCount := top100.Count - participant.StartProblemCount
	if newProblemCount < 0 {
		newProblemCount = 0
	}

	return models.ScoreData{
		ParticipantID: participant.ID,
		Name:          participant.Name,
		BaekjoonID:    participant.BaekjoonID,
		Score:         score,
		CurrentTier:   userInfo.Tier,
		CurrentRating: userInfo.Rating,
		ProblemCount:  newProblemCount,
	}, nil
}

// sortScores 점수 데이터를 정렬합니다
func (sm *ScoreboardManager) sortScores(scores []models.ScoreData) {
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
}

// groupScoresByLeague 참가자들을 시작 티어 기준으로 리그별로 분류합니다
func (sm *ScoreboardManager) groupScoresByLeague(scores []models.ScoreData) map[int][]models.ScoreData {
	leagueScores := make(map[int][]models.ScoreData)

	// 각 참가자의 시작 티어를 가져와서 리그별로 분류
	participants := sm.storage.GetParticipants()
	participantTiers := make(map[string]int)

	for _, p := range participants {
		participantTiers[p.BaekjoonID] = p.StartTier
	}

	for _, score := range scores {
		startTier, exists := participantTiers[score.BaekjoonID]
		if !exists {
			continue // 참가자 정보가 없으면 스킵
		}

		league := sm.calculator.GetUserLeague(startTier)
		leagueScores[league] = append(leagueScores[league], score)
	}

	// 각 리그별로 점수 순으로 정렬
	for league := range leagueScores {
		sort.Slice(leagueScores[league], func(i, j int) bool {
			return leagueScores[league][i].Score > leagueScores[league][j].Score
		})
	}

	return leagueScores
}

func (sm *ScoreboardManager) formatScoreboard(competition *models.Competition, scores []models.ScoreData, isAdmin bool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf(constants.MsgScoreboardTitle, competition.Name),
		Description: fmt.Sprintf("%s ~ %s",
			competition.StartDate.Format(constants.DateFormat),
			competition.EndDate.Format(constants.DateFormat)),
		Color: constants.ColorTierGold,
	}

	if len(scores) == 0 {
		embed.Description += "\n\n" + constants.MsgScoreboardNoScores
		return embed
	}

	// 리그별로 참가자들을 분류
	leagueScores := sm.groupScoresByLeague(scores)

	var sb strings.Builder

	// 각 리그별로 스코어보드 생성
	leagueOrder := []int{constants.LeagueRookie, constants.LeaguePro, constants.LeagueMax}

	for _, league := range leagueOrder {
		if leagueScores[league] == nil || len(leagueScores[league]) == 0 {
			continue
		}

		// 리그명 추가
		leagueName := sm.calculator.GetLeagueName(league)
		sb.WriteString(fmt.Sprintf("\n**🏆 %s 리그**\n", leagueName))
		sb.WriteString("```\n")
		sb.WriteString(fmt.Sprintf("%-*s %-*s %*s\n",
			constants.ScoreboardRankWidth, "순위",
			constants.ScoreboardNameWidth, "아이디",
			constants.ScoreboardScoreWidth, "점수"))
		sb.WriteString(constants.ScoreboardSeparator + "\n")

		// 해당 리그 참가자들만 표시
		for i, score := range leagueScores[league] {
			rank := i + 1
			sb.WriteString(fmt.Sprintf("%-*d  %-*s %*.0f\n",
				constants.ScoreboardRankWidth, rank,
				constants.ScoreboardNameWidth, utils.TruncateString(score.BaekjoonID, constants.ScoreboardNameWidth),
				constants.ScoreboardScoreWidth, score.Score))
		}
		sb.WriteString("```\n")
	}

	embed.Description += sb.String()

	// 블랙아웃 경고 추가
	now := utils.GetCurrentTimeKST()
	if now.Before(competition.BlackoutStartDate) {
		daysLeft := int(competition.BlackoutStartDate.Sub(now).Hours() / 24)
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf(constants.MsgScoreboardBlackoutWarning, daysLeft),
		}
	}

	return embed
}

func (sm *ScoreboardManager) SendDailyScoreboard(session *discordgo.Session, channelID string) error {
	embed, err := sm.GenerateScoreboard(false) // 자동 스코어보드는 관리자 권한 없음
	if err != nil {
		return err
	}

	_, err = session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		utils.Error("DISCORD API ERROR: Failed to send daily scoreboard: %v", err)
	}
	return err
}
