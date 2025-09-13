package storage

import (
	"discord-bot/constants"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/utils"
	"fmt"
	"sync"
	"time"
)

// InMemoryStorage 테스트/개발용 비영구 저장소 구현
type InMemoryStorage struct {
	mu           sync.RWMutex
	apiClient    interfaces.APIClient
	competition  *models.Competition
	participants map[string]models.Participant // key: BaekjoonID
}

// NewInMemoryStorage 새 인메모리 저장소 생성
func NewInMemoryStorage(apiClient interfaces.APIClient) *InMemoryStorage {
	return &InMemoryStorage{
		apiClient:    apiClient,
		participants: make(map[string]models.Participant),
	}
}

// AddParticipant 참가자 추가
func (s *InMemoryStorage) AddParticipant(name, baekjoonID string, startTier, startRating int, organizationID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !utils.IsValidUsername(name) {
		return fmt.Errorf("invalid username: %s", name)
	}
	if !utils.IsValidBaekjoonID(baekjoonID) {
		return fmt.Errorf("invalid Baekjoon ID: %s", baekjoonID)
	}
	if s.competition == nil || !s.competition.IsActive {
		return fmt.Errorf("no active competition to add participant to")
	}
	if _, exists := s.participants[baekjoonID]; exists {
		return fmt.Errorf("participant with Baekjoon ID %s already exists", baekjoonID)
	}
	for _, p := range s.participants {
		if p.Name == name {
			return fmt.Errorf("participant with name %s already exists", name)
		}
	}

	startProblemIDs, startProblemCount := s.fetchStartingProblems(baekjoonID)

	p := models.Participant{
		ID:                baekjoonID,
		Name:              utils.SanitizeString(name),
		BaekjoonID:        baekjoonID,
		OrganizationID:    organizationID,
		StartTier:         startTier,
		StartRating:       startRating,
		CreatedAt:         time.Now(),
		StartProblemIDs:   startProblemIDs,
		StartProblemCount: startProblemCount,
	}
	s.participants[baekjoonID] = p
	return nil
}

// GetParticipants 참가자 전체 조회
func (s *InMemoryStorage) GetParticipants() []models.Participant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]models.Participant, 0, len(s.participants))
	for _, p := range s.participants {
		res = append(res, p)
	}
	return res
}

// RemoveParticipant 참가자 제거
func (s *InMemoryStorage) RemoveParticipant(baekjoonID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.participants[baekjoonID]; !ok {
		return fmt.Errorf("participant not found: %s", baekjoonID)
	}
	delete(s.participants, baekjoonID)
	return nil
}

// CreateCompetition 새 대회 생성 및 활성화
func (s *InMemoryStorage) CreateCompetition(name string, startDate, endDate time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 이전 대회 비활성화
	if s.competition != nil {
		s.competition.IsActive = false
	}
	comp := &models.Competition{
		ID:                fmt.Sprintf("mem-%d", time.Now().UnixNano()),
		Name:              name,
		StartDate:         startDate,
		EndDate:           endDate,
		BlackoutStartDate: endDate.AddDate(0, 0, -constants.BlackoutDays),
		IsActive:          true,
		ShowScoreboard:    true,
	}
	s.competition = comp
	return nil
}

// GetCompetition 활성 대회 조회
func (s *InMemoryStorage) GetCompetition() *models.Competition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.competition == nil || !s.competition.IsActive {
		return nil
	}
	// 사본 반환
	c := *s.competition
	return &c
}

// SetScoreboardVisibility 스코어보드 가시성 설정
func (s *InMemoryStorage) SetScoreboardVisibility(visible bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.competition == nil || !s.competition.IsActive {
		return fmt.Errorf("no active competition to update")
	}
	s.competition.ShowScoreboard = visible
	return nil
}

// UpdateCompetitionName 이름 변경
func (s *InMemoryStorage) UpdateCompetitionName(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.competition == nil || !s.competition.IsActive {
		return fmt.Errorf("no active competition to update")
	}
	s.competition.Name = name
	return nil
}

// UpdateCompetitionStartDate 시작일 변경
func (s *InMemoryStorage) UpdateCompetitionStartDate(startDate time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.competition == nil || !s.competition.IsActive {
		return fmt.Errorf("no active competition to update")
	}
	s.competition.StartDate = startDate
	return nil
}

// UpdateCompetitionEndDate 종료일 변경 및 블랙아웃 재계산
func (s *InMemoryStorage) UpdateCompetitionEndDate(endDate time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.competition == nil || !s.competition.IsActive {
		return fmt.Errorf("no active competition to update")
	}
	s.competition.EndDate = endDate
	s.competition.BlackoutStartDate = endDate.AddDate(0, 0, -constants.BlackoutDays)
	return nil
}

// IsBlackoutPeriod 블랙아웃 기간 여부
func (s *InMemoryStorage) IsBlackoutPeriod() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.competition == nil {
		return false
	}
	now := time.Now()
	return now.After(s.competition.BlackoutStartDate) && now.Before(s.competition.EndDate)
}

// SaveCompetition no-op
func (s *InMemoryStorage) SaveCompetition() error { return nil }

// SaveParticipants no-op
func (s *InMemoryStorage) SaveParticipants() error { return nil }

// 내부 헬퍼: 시작 문제 목록 로딩
func (s *InMemoryStorage) fetchStartingProblems(baekjoonID string) ([]int, int) {
	ids := []int{}
	count := 0
	top100, err := s.apiClient.GetUserTop100(baekjoonID)
	if err == nil {
		for _, item := range top100.Items {
			ids = append(ids, item.ProblemID)
		}
		count = len(ids)
		utils.Info("Loaded %d starting problems for participant %s", count, baekjoonID)
	} else {
		utils.Warn("Failed to load starting problems for participant %s: %v", baekjoonID, err)
	}
	return ids, count
}
