package bot

import (
	"discord-bot/constants"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/performance"
	"discord-bot/utils"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ScoreboardManager struct {
	storage            interfaces.StorageRepository
	calculator         interfaces.ScoreCalculator
	client             interfaces.APIClient
	tierManager        *models.TierManager
	concurrencyManager *performance.AdaptiveConcurrencyManager
}

func NewScoreboardManager(storage interfaces.StorageRepository, calculator interfaces.ScoreCalculator, client interfaces.APIClient, tierManager *models.TierManager) *ScoreboardManager {
	return &ScoreboardManager{
		storage:            storage,
		calculator:         calculator,
		client:             client,
		tierManager:        tierManager,
		concurrencyManager: performance.NewAdaptiveConcurrencyManager(),
	}
}

func (sm *ScoreboardManager) GetStorage() interfaces.StorageRepository {
	return sm.storage
}

func (sm *ScoreboardManager) GenerateScoreboard(isAdmin bool) (*discordgo.MessageEmbed, error) {
	utils.Info("GenerateScoreboard started for admin: %t", isAdmin)
	
	competition := sm.storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		utils.Error("No active competition found")
		return nil, fmt.Errorf("활성화된 대회가 없습니다")
	}
	utils.Info("Competition found: %s, Active: %t", competition.Name, competition.IsActive)

	// 블랙아웃 체크
	if embed := sm.checkBlackoutPeriod(competition, isAdmin); embed != nil {
		utils.Info("Blackout period detected, returning blackout message")
		return embed, nil
	}
	utils.Info("Blackout check passed")

	// 참가자 체크
	participants := sm.storage.GetParticipants()
	if embed := sm.checkEmptyParticipants(competition, participants); embed != nil {
		utils.Info("No participants found, returning empty message")
		return embed, nil
	}
	utils.Info("Participants found: %d", len(participants))

	// 점수 데이터 수집
	utils.Info("About to call collectScoreData")
	scores, err := sm.collectScoreData(participants)
	utils.Info("collectScoreData call returned, checking error")
	if err != nil {
		utils.Error("Failed to collect score data: %v", err)
		return nil, err
	}
	utils.Info("Score data collected: %d scores", len(scores))

	// 포맷팅
	embed := sm.formatScoreboard(competition, scores, isAdmin)
	utils.Info("Scoreboard formatted successfully")
	return embed, nil
}

// checkBlackoutPeriod 블랙아웃 기간인지 확인하고 해당 embed 반환
func (sm *ScoreboardManager) checkBlackoutPeriod(competition *models.Competition, isAdmin bool) *discordgo.MessageEmbed {
	if sm.storage.IsBlackoutPeriod() && !isAdmin {
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
	utils.Info("collectScoreData started with %d participants", len(participants))
	
	if len(participants) == 0 {
		utils.Info("No participants, returning empty slice")
		return []models.ScoreData{}, nil
	}

	// 메모리 풀에서 재사용 가능한 리소스 가져오기
	utils.Info("Getting resources from memory pool")
	scoresPtr := performance.GetScoreDataSlice()
	defer func() {
		utils.Info("Returning ScoreDataSlice to pool")
		performance.PutScoreDataSlice(scoresPtr)
		utils.Info("ScoreDataSlice returned to pool")
	}()
	scores := *scoresPtr
	utils.Info("Memory pool resources acquired")
	
	scoreChan := performance.GetScoreDataChannel(len(participants))
	defer func() {
		utils.Info("Returning ScoreDataChannel to pool")
		performance.PutScoreDataChannel(scoreChan)
		utils.Info("ScoreDataChannel returned to pool")
	}()
	
	semaphore := performance.GetSemaphoreChannel(sm.concurrencyManager.GetCurrentLimit())
	defer func() {
		utils.Info("Returning SemaphoreChannel to pool")
		performance.PutSemaphoreChannel(semaphore)
		utils.Info("SemaphoreChannel returned to pool")
	}()
	
	var wg sync.WaitGroup
	var errorCount int64

	for _, participant := range participants {
		wg.Add(1)
		go func(p models.Participant) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startTime := time.Now()
			scoreData, err := sm.calculateParticipantScore(p)
			responseTime := time.Since(startTime)

			// 응답 시간을 적응형 동시성 관리자에 기록
			sm.concurrencyManager.RecordResponseTime(responseTime)

			if err != nil {
				utils.Warn("Failed to calculate score for participant %s: %v", p.Name, err)
				atomic.AddInt64(&errorCount, 1)
				return
			}
			scoreChan <- scoreData
		}(participant)
	}

	wg.Wait()
	close(scoreChan)

	for score := range scoreChan {
		scores = append(scores, score)
	}

	if errorCount > 0 {
		utils.Warn("Failed to calculate scores for %d participants", errorCount)
	}

	utils.Info("Successfully calculated scores for %d out of %d participants", len(scores), len(participants))
	
	// 결과 복사본 생성 (메모리 풀의 슬라이스는 재사용되므로)
	utils.Info("Creating result copy with %d scores", len(scores))
	result := make([]models.ScoreData, len(scores))
	copy(result, scores)
	utils.Info("collectScoreData completed successfully with %d scores", len(result))
	utils.Info("About to return from collectScoreData")
	return result, nil
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

	rawScore := sm.calculator.CalculateScoreWithTop100(top100, participant.StartTier, participant.StartProblemIDs)
	roundedScore := math.Round(rawScore)

	newProblemCount := top100.Count - participant.StartProblemCount
	if newProblemCount < 0 {
		newProblemCount = 0
	}

	return models.ScoreData{
		ParticipantID: participant.ID,
		Name:          participant.Name,
		BaekjoonID:    participant.BaekjoonID,
		Score:         roundedScore,
		RawScore:      rawScore,
		League:        sm.calculator.GetUserLeague(participant.StartTier),
		CurrentTier:   userInfo.Tier,
		CurrentRating: userInfo.Rating,
		ProblemCount:  newProblemCount,
	}, nil
}

// groupScoresByLeague 참가자들을 리그별로 분류하고 점수 순으로 정렬합니다
func (sm *ScoreboardManager) groupScoresByLeague(scores []models.ScoreData) map[int][]models.ScoreData {
	leagueScores := make(map[int][]models.ScoreData)

	for _, score := range scores {
		leagueScores[score.League] = append(leagueScores[score.League], score)
	}

	// 각 리그별로 점수 순으로 정렬
	for league := range leagueScores {
		sort.Slice(leagueScores[league], func(i, j int) bool {
			// 1. RawScore 기준 내림차순
			if leagueScores[league][i].RawScore != leagueScores[league][j].RawScore {
				return leagueScores[league][i].RawScore > leagueScores[league][j].RawScore
			}
			// 2. 동점일 경우 BaekjoonID 오름차순
			return leagueScores[league][i].BaekjoonID < leagueScores[league][j].BaekjoonID
		})
	}

	return leagueScores
}

// formatScoreboard 점수 데이터를 포맷팅하여 Discord 임베드 메시지로 반환합니다
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

	leagueScores := sm.groupScoresByLeague(scores)

	var sb strings.Builder

	leagueOrder := []int{constants.LeagueRookie, constants.LeaguePro, constants.LeagueMax}

	for _, league := range leagueOrder {
		if len(leagueScores[league]) == 0 {
			continue
		}

		leagueName := sm.calculator.GetLeagueName(league)
		sb.WriteString(fmt.Sprintf("\n**🏆 %s 리그**\n", leagueName))
		sb.WriteString("```\n")
		sb.WriteString(fmt.Sprintf("%-*s %-*s %*s\n",
			constants.ScoreboardRankWidth, "순위",
			constants.ScoreboardNameWidth, "아이디",
			constants.ScoreboardScoreWidth, "점수"))
		sb.WriteString(constants.ScoreboardSeparator + "\n")

		var lastRawScore float64 = -1.0
		var rank int
		for i, score := range leagueScores[league] {
			if score.RawScore != lastRawScore {
				rank = i + 1
			}
			sb.WriteString(fmt.Sprintf("%-*d  %-*s %*.0f\n",
				constants.ScoreboardRankWidth, rank,
				constants.ScoreboardNameWidth, utils.TruncateString(score.BaekjoonID, constants.ScoreboardNameWidth),
				constants.ScoreboardScoreWidth, score.Score))
			lastRawScore = score.RawScore
		}
		sb.WriteString("```\n")
	}

	embed.Description += sb.String()

	now := utils.GetCurrentTimeKST()
	if now.Before(competition.BlackoutStartDate) {
		daysLeft := int(competition.BlackoutStartDate.Sub(now).Hours() / 24)
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf(constants.MsgScoreboardBlackoutWarning, daysLeft),
		}
	}

	return embed
}

// SendDailyScoreboard 매일 스코어보드를 지정된 채널에 전송합니다
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
