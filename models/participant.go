package models

import (
	"time"
)

type Participant struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	BaekjoonID        string    `json:"baekjoon_id"`
	StartTier         int       `json:"start_tier"`
	StartRating       int       `json:"start_rating"`
	CreatedAt         time.Time `json:"created_at"`
	StartProblemIDs   []int     `json:"start_problem_ids"`   // 참가 시점의 해결한 문제 ID들
	StartProblemCount int       `json:"start_problem_count"` // 참가 시점의 해결한 문제 수
}

type Competition struct {
	ID                int           `json:"id"`
	Name              string        `json:"name"`
	StartDate         time.Time     `json:"start_date"`
	EndDate           time.Time     `json:"end_date"`
	BlackoutStartDate time.Time     `json:"blackout_start_date"`
	IsActive          bool          `json:"is_active"`
	ShowScoreboard    bool          `json:"show_scoreboard"`
	Participants      []Participant `json:"participants"`
}

type ScoreData struct {
	ParticipantID int     `json:"participant_id"`
	Name          string  `json:"name"`
	BaekjoonID    string  `json:"baekjoon_id"`
	Score         float64 `json:"score"`         // 반올림된 점수 (표시용)
	RawScore      float64 `json:"raw_score"`      // 원본 점수 (순위 결정용)
	League        int     `json:"league"`         // 참가자 리그
	CurrentTier   int     `json:"current_tier"`
	CurrentRating int     `json:"current_rating"`
	ProblemCount  int     `json:"problem_count"`
}
