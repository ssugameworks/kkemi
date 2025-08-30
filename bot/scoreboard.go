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
	"time"

	"github.com/bwmarrin/discordgo"
)

type ScoreboardManager struct {
	storage    interfaces.StorageRepository
	calculator interfaces.ScoreCalculator
	client     interfaces.APIClient
}

func NewScoreboardManager(storage interfaces.StorageRepository, calculator interfaces.ScoreCalculator, client interfaces.APIClient) *ScoreboardManager {
	return &ScoreboardManager{
		storage:    storage,
		calculator: calculator,
		client:     client,
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
		tm := models.NewTierManager()
		return &discordgo.MessageEmbed{
			Title:       "🔒 스코어보드 비공개",
			Description: "마지막 3일간 스코어보드가 비공개됩니다",
			Color:       tm.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// checkEmptyParticipants 참가자가 없는지 확인하고 해당 embed 반환
func (sm *ScoreboardManager) checkEmptyParticipants(competition *models.Competition, participants []models.Participant) *discordgo.MessageEmbed {
	if len(participants) == 0 {
		tm := models.NewTierManager()
		return &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("🏆 %s 스코어보드", competition.Name),
			Description: "참가자가 없습니다.",
			Color:       tm.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// collectScoreData 참가자들의 점수 데이터를 병렬로 수집합니다
func (sm *ScoreboardManager) collectScoreData(participants []models.Participant) ([]models.ScoreData, error) {
	if len(participants) == 0 {
		return []models.ScoreData{}, nil
	}

	// 병렬 처리를 위한 채널과 대기 그룹
	scoreChan := make(chan models.ScoreData, len(participants))
	errorChan := make(chan error, len(participants))
	semaphore := make(chan struct{}, constants.MaxConcurrentRequests)
	var wg sync.WaitGroup

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
				errorChan <- err
				return
			}
			scoreChan <- scoreData
		}(participant)
	}

	// 고루틴들이 완료될 때까지 대기
	wg.Wait()
	close(scoreChan)
	close(errorChan)

	// 결과 수집
	var scores []models.ScoreData
	for score := range scoreChan {
		scores = append(scores, score)
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

	score, err := sm.calculator.CalculateScore(participant.BaekjoonID, participant.StartTier, participant.StartProblemIDs)
	if err != nil {
		return models.ScoreData{}, err
	}

	top100, err := sm.client.GetUserTop100(participant.BaekjoonID)
	if err != nil {
		return models.ScoreData{}, err
	}

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

func (sm *ScoreboardManager) formatScoreboard(competition *models.Competition, scores []models.ScoreData, isAdmin bool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏆 %s 스코어보드", competition.Name),
		Description: fmt.Sprintf("%s ~ %s",
			competition.StartDate.Format(constants.DateFormat),
			competition.EndDate.Format(constants.DateFormat)),
		Color: constants.ColorTierGold,
	}

	if len(scores) == 0 {
		embed.Description += "\n\n아직 점수가 계산된 참가자가 없습니다."
		return embed
	}

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-*s %-*s %*s\n",
		constants.ScoreboardRankWidth, "순위", 
		constants.ScoreboardNameWidth, "이름", 
		constants.ScoreboardScoreWidth, "점수"))
	sb.WriteString(constants.ScoreboardSeparator + "\n")

	for i, score := range scores {
		rank := i + 1
		sb.WriteString(fmt.Sprintf("%-*d %-*s %*.0f\n",
			constants.ScoreboardRankWidth, rank,
			constants.ScoreboardNameWidth, utils.TruncateString(score.Name, constants.ScoreboardNameWidth),
			constants.ScoreboardScoreWidth, score.Score))
	}

	sb.WriteString("```")

	embed.Description += "\n\n" + sb.String()

	// 블랙아웃 경고 추가
	now := time.Now()
	if now.Before(competition.BlackoutStartDate) {
		daysLeft := int(competition.BlackoutStartDate.Sub(now).Hours() / 24)
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("⚠️ %d일 후 스코어보드가 비공개됩니다", daysLeft),
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
	return err
}
