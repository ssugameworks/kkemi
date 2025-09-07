package storage

import (
	"context"
	"discord-bot/constants"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/utils"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// FirebaseStorage 참가자와 대회 데이터를 Firestore에서 관리하는 저장소입니다
type FirebaseStorage struct {
	client    *firestore.Client
	apiClient interfaces.APIClient
	ctx       context.Context
}

// NewStorage 새로운 FirebaseStorage 인스턴스를 생성하고 Firestore에 연결합니다
func NewStorage(apiClient interfaces.APIClient) (interfaces.StorageRepository, error) {
	utils.Info("Initializing Firebase storage system")
	ctx := context.Background()

	// Railway 환경 변수에서 직접 JSON 내용을 읽어옵니다.
	firebaseCreds := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if firebaseCreds == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS_JSON environment variable not set")
	}

	opt := option.WithCredentialsJSON([]byte(firebaseCreds))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firestore client: %v", err)
	}

	s := &FirebaseStorage{
		client:    client,
		apiClient: apiClient,
		ctx:       ctx,
	}

	utils.Info("Firebase storage system initialized successfully")
	return s, nil
}

// AddParticipant 새로운 참가자를 Firestore에 추가합니다
func (s *FirebaseStorage) AddParticipant(name, baekjoonID string, startTier, startRating int) error {
	competition := s.GetCompetition()
	if competition == nil {
		return fmt.Errorf("no active competition to add participant to")
	}

	startProblemIDs, startProblemCount := s.fetchStartingProblems(baekjoonID)

	participant := models.Participant{
		Name:              utils.SanitizeString(name),
		BaekjoonID:        baekjoonID,
		StartTier:         startTier,
		StartRating:       startRating,
		CreatedAt:         time.Now(),
		StartProblemIDs:   startProblemIDs,
		StartProblemCount: startProblemCount,
	}

	// 참가자 ID를 백준 ID로 사용하여 중복 방지 및 조회 용이성 확보
	_, err := s.client.Collection("competitions").Doc(competition.ID).Collection("participants").Doc(baekjoonID).Set(s.ctx, participant)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	utils.Info("Added new participant to Firestore: %s (%s)", name, baekjoonID)
	return nil
}

// GetParticipants 현재 대회에 등록된 모든 참가자를 Firestore에서 조회합니다
func (s *FirebaseStorage) GetParticipants() []models.Participant {
	competition := s.GetCompetition()
	if competition == nil {
		return []models.Participant{}
	}

	var participants []models.Participant
	iter := s.client.Collection("competitions").Doc(competition.ID).Collection("participants").Documents(s.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			utils.Error("Failed to iterate participants: %v", err)
			return participants
		}

		var p models.Participant
		doc.DataTo(&p)
		// Firestore 문서 ID가 백준 ID이므로 별도로 설정할 필요 없음
		participants = append(participants, p)
	}
	return participants
}

// RemoveParticipant 백준ID로 참가자를 Firestore에서 삭제합니다
func (s *FirebaseStorage) RemoveParticipant(baekjoonID string) error {
	competition := s.GetCompetition()
	if competition == nil {
		return fmt.Errorf("no active competition")
	}

	_, err := s.client.Collection("competitions").Doc(competition.ID).Collection("participants").Doc(baekjoonID).Delete(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to remove participant from Firestore: %w", err)
	}

	utils.Info("Removed participant from Firestore: %s", baekjoonID)
	return nil
}

// CreateCompetition 새로운 대회를 Firestore에 생성합니다
func (s *FirebaseStorage) CreateCompetition(name string, startDate, endDate time.Time) error {
	blackoutStart := endDate.AddDate(0, 0, -constants.BlackoutDays)

	newComp := models.Competition{
		Name:              name,
		StartDate:         startDate,
		EndDate:           endDate,
		BlackoutStartDate: blackoutStart,
		IsActive:          true,
		ShowScoreboard:    true,
	}

	// 모든 대회를 비활성화
	iter := s.client.Collection("competitions").Where("IsActive", "==", true).Documents(s.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate existing competitions: %w", err)
		}
		_, err = doc.Ref.Update(s.ctx, []firestore.Update{{Path: "IsActive", Value: false}})
		if err != nil {
			return fmt.Errorf("failed to deactivate old competition: %w", err)
		}
	}

	// 새 대회를 활성 상태로 추가
	_, _, err := s.client.Collection("competitions").Add(s.ctx, newComp)
	return err
}

// GetCompetition 현재 활성화된 대회를 Firestore에서 조회합니다
func (s *FirebaseStorage) GetCompetition() *models.Competition {
	iter := s.client.Collection("competitions").Where("IsActive", "==", true).Limit(1).Documents(s.ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil // 활성화된 대회 없음
	}
	if err != nil {
		utils.Error("Failed to get active competition: %v", err)
		return nil
	}

	var c models.Competition
	doc.DataTo(&c)
	c.ID = doc.Ref.ID // Firestore 문서 ID를 모델 ID로 사용

	return &c
}

// updateActiveCompetitionField 활성화된 대회의 특정 필드를 업데이트합니다
func (s *FirebaseStorage) updateActiveCompetitionField(updates []firestore.Update) error {
	competition := s.GetCompetition()
	if competition == nil {
		return fmt.Errorf("no active competition to update")
	}
	_, err := s.client.Collection("competitions").Doc(competition.ID).Update(s.ctx, updates)
	return err
}

// UpdateCompetitionName 대회명을 업데이트합니다
func (s *FirebaseStorage) UpdateCompetitionName(name string) error {
	return s.updateActiveCompetitionField([]firestore.Update{{Path: "Name", Value: name}})
}

// UpdateCompetitionStartDate 대회 시작일을 업데이트합니다
func (s *FirebaseStorage) UpdateCompetitionStartDate(startDate time.Time) error {
	return s.updateActiveCompetitionField([]firestore.Update{{Path: "StartDate", Value: startDate}})
}

// UpdateCompetitionEndDate 대회 종료일을 업데이트하고 블랙아웃 기간도 자동으로 재설정합니다
func (s *FirebaseStorage) UpdateCompetitionEndDate(endDate time.Time) error {
	updates := []firestore.Update{
		{Path: "EndDate", Value: endDate},
		{Path: "BlackoutStartDate", Value: endDate.AddDate(0, 0, -constants.BlackoutDays)},
	}
	return s.updateActiveCompetitionField(updates)
}

// SetScoreboardVisibility 스코어보드 공개 여부를 설정합니다
func (s *FirebaseStorage) SetScoreboardVisibility(visible bool) error {
	return s.updateActiveCompetitionField([]firestore.Update{{Path: "ShowScoreboard", Value: visible}})
}

// IsBlackoutPeriod 현재가 블랙아웃 기간인지 확인합니다
func (s *FirebaseStorage) IsBlackoutPeriod() bool {
	comp := s.GetCompetition()
	if comp == nil {
		return false
	}
	now := time.Now()
	return now.After(comp.BlackoutStartDate) && now.Before(comp.EndDate)
}

// fetchStartingProblems 참가 시점의 해결한 문제들을 가져옵니다 (이전 로직과 동일)
func (s *FirebaseStorage) fetchStartingProblems(baekjoonID string) ([]int, int) {
	startProblemIDs := []int{}
	startProblemCount := 0

	top100, err := s.apiClient.GetUserTop100(baekjoonID)
	if err == nil {
		for _, problem := range top100.Items {
			startProblemIDs = append(startProblemIDs, problem.ProblemID)
		}
		startProblemCount = len(startProblemIDs)
		utils.Info("Loaded %d starting problems for participant %s", startProblemCount, baekjoonID)
	} else {
		utils.Warn("Failed to load starting problems for participant %s: %v", baekjoonID, err)
	}

	return startProblemIDs, startProblemCount
}
