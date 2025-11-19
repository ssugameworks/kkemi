package bot

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ssugameworks/kkemi/constants"
	"github.com/ssugameworks/kkemi/interfaces"
	"github.com/ssugameworks/kkemi/models"
	"github.com/ssugameworks/kkemi/performance"
	"github.com/ssugameworks/kkemi/utils"

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

func (manager *ScoreboardManager) GetStorage() interfaces.StorageRepository {
	return manager.storage
}

func (manager *ScoreboardManager) GenerateScoreboard(isAdmin bool) (*discordgo.MessageEmbed, error) {
	competition := manager.storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		return nil, fmt.Errorf("활성화된 대회가 없습니다")
	}

	// 마지막날 체크
	now := utils.GetCurrentTimeKST()
	isLastDay := now.Year() == competition.EndDate.Year() &&
		now.Month() == competition.EndDate.Month() &&
		now.Day() == competition.EndDate.Day()

	// 블랙아웃 체크 (마지막날에는 공개)
	if embed := manager.checkBlackoutPeriod(competition, isAdmin || isLastDay); embed != nil {
		return embed, nil
	}

	// 참가자 체크
	participants := manager.storage.GetParticipants()
	if embed := manager.checkEmptyParticipants(competition, participants); embed != nil {
		return embed, nil
	}

	// 점수 데이터 수집
	scores, err := manager.collectScoreData(participants)
	if err != nil {
		return nil, err
	}

	// 포맷팅
	return manager.formatScoreboard(competition, scores, isAdmin), nil
}

// CollectScoreData 참가자들의 점수 데이터를 수집하여 반환합니다 (외부 접근용)
func (manager *ScoreboardManager) CollectScoreData() ([]models.ScoreData, error) {
	competition := manager.storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		return nil, fmt.Errorf("활성화된 대회가 없습니다")
	}

	participants := manager.storage.GetParticipants()
	if len(participants) == 0 {
		return []models.ScoreData{}, nil
	}

	return manager.collectScoreData(participants)
}

// checkBlackoutPeriod 블랙아웃 기간인지 확인하고 해당 embed 반환
func (manager *ScoreboardManager) checkBlackoutPeriod(competition *models.Competition, isAdmin bool) *discordgo.MessageEmbed {
	if manager.storage.IsBlackoutPeriod() && !isAdmin {
		return &discordgo.MessageEmbed{
			Title:       constants.MsgScoreboardBlackout,
			Description: constants.MsgScoreboardBlackoutDesc,
			Color:       manager.tierManager.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// checkEmptyParticipants 참가자가 없는지 확인하고 해당 embed 반환
func (manager *ScoreboardManager) checkEmptyParticipants(competition *models.Competition, participants []models.Participant) *discordgo.MessageEmbed {
	if len(participants) == 0 {
		return &discordgo.MessageEmbed{
			Title:       fmt.Sprintf(constants.MsgScoreboardTitle, competition.Name),
			Description: constants.MsgScoreboardNoParticipants,
			Color:       manager.tierManager.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// collectScoreData 참가자들의 점수 데이터를 병렬로 수집합니다
func (manager *ScoreboardManager) collectScoreData(participants []models.Participant) ([]models.ScoreData, error) {
	if len(participants) == 0 {
		return []models.ScoreData{}, nil
	}

	// 메모리 풀에서 재사용 가능한 리소스 가져오기
	scoresPtr := performance.GetScoreDataSlice()
	defer performance.PutScoreDataSlice(scoresPtr)
	scores := *scoresPtr

	scoreChan := performance.GetScoreDataChannel(len(participants))
	defer performance.PutScoreDataChannel(scoreChan)

	semaphore := performance.GetSemaphoreChannel(manager.concurrencyManager.GetCurrentLimit())
	defer performance.PutSemaphoreChannel(semaphore)

	var wg sync.WaitGroup
	var errorCount int64

	for _, participant := range participants {
		wg.Add(1)
		go func(p models.Participant) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startTime := time.Now()
			scoreData, err := manager.calculateParticipantScore(p)
			responseTime := time.Since(startTime)

			// 응답 시간을 적응형 동시성 관리자에 기록
			manager.concurrencyManager.RecordResponseTime(responseTime)

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
	result := make([]models.ScoreData, len(scores))
	copy(result, scores)
	return result, nil
}

// calculateParticipantScore 개별 참가자의 점수를 계산합니다
func (manager *ScoreboardManager) calculateParticipantScore(participant models.Participant) (models.ScoreData, error) {
	ctx := context.Background()
	userInfo, err := manager.client.GetUserInfo(ctx, participant.BaekjoonID)
	if err != nil {
		return models.ScoreData{}, err
	}

	top100, err := manager.client.GetUserTop100(ctx, participant.BaekjoonID)
	if err != nil {
		return models.ScoreData{}, err
	}

	rawScore := manager.calculator.CalculateScoreWithTop100(top100, participant.StartTier, participant.StartProblemIDs)
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
		League:        manager.calculator.GetUserLeague(participant.StartTier),
		CurrentTier:   userInfo.Tier,
		CurrentRating: userInfo.Rating,
		ProblemCount:  newProblemCount,
	}, nil
}

// groupScoresByLeague 참가자들을 리그별로 분류하고 점수 순으로 정렬합니다
func (manager *ScoreboardManager) groupScoresByLeague(scores []models.ScoreData) map[int][]models.ScoreData {
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
func (manager *ScoreboardManager) formatScoreboard(competition *models.Competition, scores []models.ScoreData, isAdmin bool) *discordgo.MessageEmbed {
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

	leagueScores := manager.groupScoresByLeague(scores)

	var builder strings.Builder

	leagueOrder := []int{constants.LeagueRookie, constants.LeaguePro, constants.LeagueMaster}

	for _, league := range leagueOrder {
		if len(leagueScores[league]) == 0 {
			continue
		}

		leagueName := manager.calculator.GetLeagueName(league)
		builder.WriteString(fmt.Sprintf("\n**🏆 %s 리그**\n", leagueName))
		builder.WriteString("```\n")
		builder.WriteString(fmt.Sprintf("%-*s %-*s %*s\n",
			constants.ScoreboardRankWidth, "순위",
			constants.ScoreboardNameWidth, "아이디",
			constants.ScoreboardScoreWidth, "점수"))
		builder.WriteString(constants.ScoreboardSeparator + "\n")

		var lastRawScore float64 = -1.0
		var rank int
		for i, score := range leagueScores[league] {
			if score.RawScore != lastRawScore {
				rank = i + 1
			}
			builder.WriteString(fmt.Sprintf("%-*d  %-*s %*.0f\n",
				constants.ScoreboardRankWidth, rank,
				constants.ScoreboardNameWidth, utils.TruncateString(score.BaekjoonID, constants.ScoreboardNameWidth),
				constants.ScoreboardScoreWidth, score.Score))
			lastRawScore = score.RawScore
		}
		builder.WriteString("```\n")
	}

	embed.Description += builder.String()

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
func (manager *ScoreboardManager) SendDailyScoreboard(session *discordgo.Session, channelID string) error {
	embed, err := manager.GenerateScoreboard(false) // 자동 스코어보드는 관리자 권한 없음
	if err != nil {
		return err
	}

	_, err = session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		utils.Error("DISCORD API ERROR: Failed to send daily scoreboard: %v", err)
	}
	return err
}
