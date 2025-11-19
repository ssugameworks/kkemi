package interfaces

import (
	"context"

	"github.com/ssugameworks/kkemi/api"
)

// ScoreCalculator 점수 계산을 위한 인터페이스입니다
type ScoreCalculator interface {
	CalculateScore(ctx context.Context, handle string, startTier int, startProblemIDs []int) (float64, error)
	CalculateScoreWithTop100(top100 *api.Top100Response, startTier int, startProblemIDs []int) float64
	GetUserLeague(startTier int) int
	GetLeagueName(league int) string
}
