package storage

import (
	"context"
	"discord-bot/constants"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/utils"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// FirebaseStorage Firestore를 사용하여 데이터를 관리하는 저장소입니다.
type FirebaseStorage struct {
	client         *firestore.Client
	apiClient      interfaces.APIClient
	ctx            context.Context
	app            *firebase.App
	reconnectMutex sync.Mutex
}

// 에러 복구 관련 상수
const (
	maxReconnectAttempts = 3
	reconnectDelay       = 2 * time.Second
)

// NewStorage 새로운 FirebaseStorage 인스턴스를 생성하고 Firestore에 연결합니다.
func NewStorage(apiClient interfaces.APIClient) (interfaces.StorageRepository, error) {
	utils.Info("Initializing Firebase storage system")
	ctx := context.Background()

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
		app:       app,
	}

	utils.Info("Firebase storage system initialized successfully")
	return s, nil
}

// GetClient Firestore 클라이언트를 반환합니다 (헬스체크용)
func (s *FirebaseStorage) GetClient() interface{} {
	return s.client
}

// reconnectFirestore Firestore 클라이언트를 재연결합니다
func (s *FirebaseStorage) reconnectFirestore() error {
	s.reconnectMutex.Lock()
	defer s.reconnectMutex.Unlock()

	utils.Warn("Attempting to reconnect to Firestore")

	for attempt := 1; attempt <= maxReconnectAttempts; attempt++ {
		// 기존 클라이언트 종료
		if s.client != nil {
			s.client.Close()
		}

		// 새 클라이언트 생성
		newClient, err := s.app.Firestore(s.ctx)
		if err != nil {
			utils.Warn("Firestore reconnection attempt %d/%d failed: %v", attempt, maxReconnectAttempts, err)
			if attempt < maxReconnectAttempts {
				time.Sleep(reconnectDelay * time.Duration(attempt)) // 점진적 지연
			}
			continue
		}

		s.client = newClient
		utils.Info("Successfully reconnected to Firestore on attempt %d", attempt)
		return nil
	}

	return fmt.Errorf("failed to reconnect to Firestore after %d attempts", maxReconnectAttempts)
}

// executeWithRetry Firestore 작업을 재시도 로직과 함께 실행합니다
func (s *FirebaseStorage) executeWithRetry(operation func() error) error {
	err := operation()
	if err != nil {
		// Firestore 연결 오류인 경우 재연결 시도
		if isFirestoreConnectionError(err) {
			utils.Warn("Detected Firestore connection error, attempting reconnection: %v", err)
			if reconnectErr := s.reconnectFirestore(); reconnectErr != nil {
				return fmt.Errorf("operation failed and reconnection failed: %v (original: %v)", reconnectErr, err)
			}
			// 재연결 성공 시 작업 재시도
			return operation()
		}
	}
	return err
}

// isFirestoreConnectionError Firestore 연결 관련 에러인지 확인합니다
func isFirestoreConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return fmt.Sprintf("%v", err) != "" && (
	// 일반적인 연결 오류 패턴들
	contains(errStr, "connection") ||
		contains(errStr, "network") ||
		contains(errStr, "timeout") ||
		contains(errStr, "unavailable") ||
		contains(errStr, "deadline exceeded"))
}

// contains 문자열 포함 여부를 확인하는 헬퍼 함수
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring 부분 문자열을 찾는 헬퍼 함수
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// AddParticipant 새로운 참가자를 Firestore에 추가합니다.
func (s *FirebaseStorage) AddParticipant(name, baekjoonID string, startTier, startRating int, organizationID int) error {
	return s.executeWithRetry(func() error {
		// 입력값 검증
		if !utils.IsValidUsername(name) {
			return fmt.Errorf("invalid username: %s", name)
		}
		if !utils.IsValidBaekjoonID(baekjoonID) {
			return fmt.Errorf("invalid Baekjoon ID: %s", baekjoonID)
		}

		competition := s.GetCompetition()
		if competition == nil {
			return fmt.Errorf("no active competition to add participant to")
		}

		// 중복 확인
		existingDoc, err := s.client.Collection("competitions").Doc(competition.ID).Collection("participants").Doc(baekjoonID).Get(s.ctx)
		if err == nil && existingDoc.Exists() {
			return fmt.Errorf("participant with Baekjoon ID %s already exists", baekjoonID)
		}

		// 이름 중복 확인
		iter := s.client.Collection("competitions").Doc(competition.ID).Collection("participants").Where("name", "==", name).Limit(1).Documents(s.ctx)
		if doc, err := iter.Next(); err == nil && doc != nil {
			return fmt.Errorf("participant with name %s already exists", name)
		}

		startProblemIDs, startProblemCount := s.fetchStartingProblems(baekjoonID)

		participant := models.Participant{
			Name:              utils.SanitizeString(name),
			BaekjoonID:        baekjoonID,
			OrganizationID:    organizationID,
			StartTier:         startTier,
			StartRating:       startRating,
			CreatedAt:         time.Now(),
			StartProblemIDs:   startProblemIDs,
			StartProblemCount: startProblemCount,
		}

		_, err = s.client.Collection("competitions").Doc(competition.ID).Collection("participants").Doc(baekjoonID).Set(s.ctx, participant)
		if err != nil {
			return fmt.Errorf("failed to add participant: %w", err)
		}

		utils.Info("Added new participant to Firestore: %s (%s)", name, baekjoonID)
		return nil
	})
}

// GetParticipants 현재 대회에 등록된 모든 참가자를 Firestore에서 조회합니다.
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
		p.ID = doc.Ref.ID
		participants = append(participants, p)
	}
	return participants
}

// RemoveParticipant 백준ID로 참가자를 Firestore에서 삭제합니다.
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

// CreateCompetition 새로운 대회를 Firestore에 생성합니다.
func (s *FirebaseStorage) CreateCompetition(name string, startDate, endDate time.Time) error {
	// 모든 대회를 비활성화
	iter := s.client.Collection("competitions").Where("isActive", "==", true).Documents(s.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate existing competitions: %w", err)
		}
		_, err = doc.Ref.Update(s.ctx, []firestore.Update{{Path: "isActive", Value: false}})
		if err != nil {
			return fmt.Errorf("failed to deactivate old competition: %w", err)
		}
	}

	newComp := models.Competition{
		Name:              name,
		StartDate:         startDate,
		EndDate:           endDate,
		BlackoutStartDate: endDate.AddDate(0, 0, -constants.BlackoutDays),
		IsActive:          true,
		ShowScoreboard:    true,
	}

	_, _, err := s.client.Collection("competitions").Add(s.ctx, newComp)
	return err
}

// GetCompetition 현재 활성화된 대회를 Firestore에서 조회합니다.
func (s *FirebaseStorage) GetCompetition() *models.Competition {
	iter := s.client.Collection("competitions").Where("isActive", "==", true).Limit(1).Documents(s.ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil
	}
	if err != nil {
		utils.Error("Failed to get active competition: %v", err)
		return nil
	}

	var c models.Competition
	doc.DataTo(&c)
	c.ID = doc.Ref.ID
	return &c
}

// updateActiveCompetitionField 활성화된 대회의 특정 필드를 업데이트합니다.
func (s *FirebaseStorage) updateActiveCompetitionField(updates []firestore.Update) error {
	competition := s.GetCompetition()
	if competition == nil {
		return fmt.Errorf("no active competition to update")
	}
	_, err := s.client.Collection("competitions").Doc(competition.ID).Update(s.ctx, updates)
	return err
}

func (s *FirebaseStorage) UpdateCompetitionName(name string) error {
	return s.updateActiveCompetitionField([]firestore.Update{{Path: "name", Value: name}})
}

func (s *FirebaseStorage) UpdateCompetitionStartDate(startDate time.Time) error {
	return s.updateActiveCompetitionField([]firestore.Update{{Path: "startDate", Value: startDate}})
}

func (s *FirebaseStorage) UpdateCompetitionEndDate(endDate time.Time) error {
	updates := []firestore.Update{
		{Path: "endDate", Value: endDate},
		{Path: "blackoutStartDate", Value: endDate.AddDate(0, 0, -constants.BlackoutDays)},
	}
	return s.updateActiveCompetitionField(updates)
}

func (s *FirebaseStorage) SetScoreboardVisibility(visible bool) error {
	return s.updateActiveCompetitionField([]firestore.Update{{Path: "showScoreboard", Value: visible}})
}

func (s *FirebaseStorage) IsBlackoutPeriod() bool {
	comp := s.GetCompetition()
	if comp == nil {
		return false
	}
	now := time.Now()
	return now.After(comp.BlackoutStartDate) && now.Before(comp.EndDate)
}

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

// SaveCompetition Firestore에서 쓰기 작업이 즉시 이루어지므로 no-op입니다.
func (s *FirebaseStorage) SaveCompetition() error {
	return nil
}

// SaveParticipants Firestore에서 쓰기 작업이 즉시 이루어지므로 no-op입니다.
func (s *FirebaseStorage) SaveParticipants() error {
	return nil
}
