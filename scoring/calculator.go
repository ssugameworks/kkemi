package scoring

import (
	"discord-bot/api"
	"discord-bot/constants"
	"discord-bot/interfaces"
	"discord-bot/models"
	"math"
)

type ScoreCalculator struct {
	client      interfaces.APIClient
	tierManager *models.TierManager
}

func NewScoreCalculator(apiClient interfaces.APIClient, tierManager *models.TierManager) interfaces.ScoreCalculator {
	return &ScoreCalculator{
		client:      apiClient,
		tierManager: tierManager,
	}
}

func (sc *ScoreCalculator) CalculateScore(handle string, startTier int, startProblemIDs []int) (float64, error) {
	top100, err := sc.client.GetUserTop100(handle)
	if err != nil {
		return 0, err
	}

	return sc.CalculateScoreWithTop100(top100, startTier, startProblemIDs), nil
}

func (sc *ScoreCalculator) CalculateScoreWithTop100(top100 *api.Top100Response, startTier int, startProblemIDs []int) float64 {
	// 시작 시점 문제 ID들을 맵으로 변환
	startProblemsMap := make(map[int]bool)
	for _, id := range startProblemIDs {
		startProblemsMap[id] = true
	}

	// 참가자의 리그 결정 (등록 시점 티어 기준)
	userLeague := sc.getUserLeague(startTier)

	totalScore := 0.0

	for _, problem := range top100.Items {
		// 참가 시점에 이미 해결한 문제는 제외
		if startProblemsMap[problem.ProblemID] {
			continue
		}

		problemLevel := problem.Level
		// solved.ac 난이도 값을 그대로 점수로 사용 (Bronze V = 1, Bronze IV = 2, ...)
		difficultyValue := float64(problemLevel)
		if difficultyValue == 0 {
			continue
		}

		// 새로운 가중치 계산 (리그별 + 문제 난이도 vs 시작 티어)
		weight := sc.getWeightByLeague(problemLevel, startTier, userLeague)
		score := difficultyValue * weight
		totalScore += score
	}

	// 최종 점수는 반올림하여 정수로 반환 (테스트 기대치와 일치)
	return math.Round(totalScore)
}

// getUserLeague 사용자의 등록 시점 티어를 기준으로 리그를 결정합니다
func (sc *ScoreCalculator) getUserLeague(startTier int) int {
	// 루키: Unrated ~ Silver V (티어 0-6)
	if startTier <= 6 {
		return constants.LeagueRookie
	}
	// 프로: Silver IV ~ Gold V (티어 7-11)
	if startTier <= 11 {
		return constants.LeaguePro
	}
	// 맥스: Gold IV ~ (티어 12 이상)
	return constants.LeagueMax
}

// getWeightByLeague 리그별 가중치를 계산합니다
func (sc *ScoreCalculator) getWeightByLeague(problemLevel, startTier, userLeague int) float64 {
	switch userLeague {
	case constants.LeagueRookie:
		return sc.getRookieWeight(problemLevel, startTier)
	case constants.LeaguePro:
		return sc.getProWeight(problemLevel, startTier)
	case constants.LeagueMax:
		return sc.getMaxWeight(problemLevel, startTier)
	default:
		return 1.0
	}
}

// getRookieWeight 루키 리그 가중치를 계산합니다
func (sc *ScoreCalculator) getRookieWeight(problemLevel, startTier int) float64 {
	if problemLevel > startTier {
		return constants.RookieUpperMultiplier // × 1.4
	} else if problemLevel == startTier {
		return constants.RookieBaseMultiplier // × 1.0
	} else {
		return constants.RookieLowerMultiplier // × 0.5
	}
}

// getProWeight 프로 리그 가중치를 계산합니다
func (sc *ScoreCalculator) getProWeight(problemLevel, startTier int) float64 {
	if problemLevel > startTier {
		return constants.ProUpperMultiplier // × 1.2
	} else if problemLevel == startTier {
		return constants.ProBaseMultiplier // × 1.0
	} else {
		return constants.ProLowerMultiplier // × 0.8
	}
}

// getMaxWeight 맥스 리그 가중치를 계산합니다
func (sc *ScoreCalculator) getMaxWeight(problemLevel, startTier int) float64 {
	if problemLevel > startTier {
		return constants.MaxUpperMultiplier // × 1.0
	} else if problemLevel == startTier {
		return constants.MaxBaseMultiplier // × 1.0
	} else {
		return constants.MaxLowerMultiplier // × 1.0
	}
}

// GetLeagueName 리그 번호를 리그 이름으로 변환합니다
func (sc *ScoreCalculator) GetLeagueName(league int) string {
	switch league {
	case constants.LeagueRookie:
		return "루키"
	case constants.LeaguePro:
		return "프로"
	case constants.LeagueMax:
		return "맥스"
	default:
		return "알 수 없음"
	}
}

// GetUserLeague 외부에서 사용할 수 있도록 노출합니다
func (sc *ScoreCalculator) GetUserLeague(startTier int) int {
	return sc.getUserLeague(startTier)
}
