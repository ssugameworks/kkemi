package interfaces

import (
	"time"

	"github.com/ssugameworks/kkemi/models"
)

// StorageRepository 데이터 저장소 작업을 위한 인터페이스입니다
type StorageRepository interface {
	// 참가자 작업
	GetParticipants() []models.Participant
	AddParticipant(name, baekjoonID string, startTier, startRating int, organizationID int) error
	RemoveParticipant(baekjoonID string) error
	SaveParticipants() error

	// 대회 작업
	GetCompetition() *models.Competition
	CreateCompetition(name string, startDate, endDate time.Time) error
	SetScoreboardVisibility(visible bool) error
	IsBlackoutPeriod() bool
	SaveCompetition() error
	UpdateCompetitionName(name string) error
	UpdateCompetitionStartDate(startDate time.Time) error
	UpdateCompetitionEndDate(endDate time.Time) error

	// 리소스 정리
	Close() error
}
