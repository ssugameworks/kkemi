package models

import (
	"time"
)

type Participant struct {
	ID                string    `firestore:"-"` // Firestore 문서 ID, firestore 필드에서는 제외
	Name              string    `firestore:"name"`
	BaekjoonID        string    `firestore:"baekjoonId"`
	OrganizationID    int       `firestore:"organizationId"`
	StartTier         int       `firestore:"startTier"`
	StartRating       int       `firestore:"startRating"`
	CreatedAt         time.Time `firestore:"createdAt"`
	StartProblemIDs   []int     `firestore:"startProblemIds"`
	StartProblemCount int       `firestore:"startProblemCount"`
}

type Competition struct {
	ID                string    `firestore:"-"` // Firestore 문서 ID, firestore 필드에서는 제외
	Name              string    `firestore:"name"`
	StartDate         time.Time `firestore:"startDate"`
	EndDate           time.Time `firestore:"endDate"`
	BlackoutStartDate time.Time `firestore:"blackoutStartDate"`
	IsActive          bool      `firestore:"isActive"`
	ShowScoreboard    bool      `firestore:"showScoreboard"`
}

type ScoreData struct {
	ParticipantID string  `json:"participant_id"`
	Name          string  `json:"name"`
	BaekjoonID    string  `json:"baekjoon_id"`
	Score         float64 `json:"score"`
	RawScore      float64 `json:"raw_score"`
	League        int     `json:"league"`
	CurrentTier   int     `json:"current_tier"`
	CurrentRating int     `json:"current_rating"`
	ProblemCount  int     `json:"problem_count"`
}
