package models

import (
	"testing"
	"time"
)

func TestParticipant_Creation(t *testing.T) {
	now := time.Now()
	participant := Participant{
		ID:                1,
		Name:              "테스트 사용자",
		BaekjoonID:        "testuser",
		StartTier:         11,
		StartRating:       1400,
		CreatedAt:         now,
		StartProblemIDs:   []int{1001, 1002, 1003},
		StartProblemCount: 3,
	}

	if participant.ID != 1 {
		t.Errorf("Expected ID 1, got %d", participant.ID)
	}
	if participant.Name != "테스트 사용자" {
		t.Errorf("Expected name '테스트 사용자', got '%s'", participant.Name)
	}
	if participant.BaekjoonID != "testuser" {
		t.Errorf("Expected baekjoon ID 'testuser', got '%s'", participant.BaekjoonID)
	}
	if participant.StartTier != 11 {
		t.Errorf("Expected start tier 11, got %d", participant.StartTier)
	}
	if participant.StartRating != 1400 {
		t.Errorf("Expected start rating 1400, got %d", participant.StartRating)
	}
	if !participant.CreatedAt.Equal(now) {
		t.Errorf("Expected created at %v, got %v", now, participant.CreatedAt)
	}
	if len(participant.StartProblemIDs) != 3 {
		t.Errorf("Expected 3 start problem IDs, got %d", len(participant.StartProblemIDs))
	}
	if participant.StartProblemCount != 3 {
		t.Errorf("Expected start problem count 3, got %d", participant.StartProblemCount)
	}
}

func TestCompetition_Creation(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	blackoutStartDate := time.Date(2024, 1, 28, 23, 59, 59, 0, time.UTC)

	participants := []Participant{
		{ID: 1, Name: "참가자1", BaekjoonID: "user1"},
		{ID: 2, Name: "참가자2", BaekjoonID: "user2"},
	}

	competition := Competition{
		ID:                1,
		Name:              "테스트 대회",
		StartDate:         startDate,
		EndDate:           endDate,
		BlackoutStartDate: blackoutStartDate,
		IsActive:          true,
		ShowScoreboard:    true,
		Participants:      participants,
	}

	if competition.ID != 1 {
		t.Errorf("Expected ID 1, got %d", competition.ID)
	}
	if competition.Name != "테스트 대회" {
		t.Errorf("Expected name '테스트 대회', got '%s'", competition.Name)
	}
	if !competition.StartDate.Equal(startDate) {
		t.Errorf("Expected start date %v, got %v", startDate, competition.StartDate)
	}
	if !competition.EndDate.Equal(endDate) {
		t.Errorf("Expected end date %v, got %v", endDate, competition.EndDate)
	}
	if !competition.BlackoutStartDate.Equal(blackoutStartDate) {
		t.Errorf("Expected blackout start date %v, got %v", blackoutStartDate, competition.BlackoutStartDate)
	}
	if !competition.IsActive {
		t.Error("Competition should be active")
	}
	if !competition.ShowScoreboard {
		t.Error("Scoreboard should be visible")
	}
	if len(competition.Participants) != 2 {
		t.Errorf("Expected 2 participants, got %d", len(competition.Participants))
	}
}

func TestScoreData_Creation(t *testing.T) {
	scoreData := ScoreData{
		ParticipantID: 1,
		Name:          "테스트 사용자",
		BaekjoonID:    "testuser",
		Score:         1250.5,
		CurrentTier:   12,
		CurrentRating: 1500,
		ProblemCount:  25,
	}

	if scoreData.ParticipantID != 1 {
		t.Errorf("Expected participant ID 1, got %d", scoreData.ParticipantID)
	}
	if scoreData.Name != "테스트 사용자" {
		t.Errorf("Expected name '테스트 사용자', got '%s'", scoreData.Name)
	}
	if scoreData.BaekjoonID != "testuser" {
		t.Errorf("Expected baekjoon ID 'testuser', got '%s'", scoreData.BaekjoonID)
	}
	if scoreData.Score != 1250.5 {
		t.Errorf("Expected score 1250.5, got %f", scoreData.Score)
	}
	if scoreData.CurrentTier != 12 {
		t.Errorf("Expected current tier 12, got %d", scoreData.CurrentTier)
	}
	if scoreData.CurrentRating != 1500 {
		t.Errorf("Expected current rating 1500, got %d", scoreData.CurrentRating)
	}
	if scoreData.ProblemCount != 25 {
		t.Errorf("Expected problem count 25, got %d", scoreData.ProblemCount)
	}
}