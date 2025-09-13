package bot

import (
	"discord-bot/interfaces"
	"discord-bot/models"
)

// CommandDependencies 명령어 핸들러가 필요로 하는 모든 의존성을 묶어서 관리합니다
type CommandDependencies struct {
	Storage           interfaces.StorageRepository
	APIClient         interfaces.APIClient
	ScoreboardManager *ScoreboardManager
	TierManager       *models.TierManager
	ScoreCalculator   interfaces.ScoreCalculator
}

// NewCommandDependencies 새로운 CommandDependencies 인스턴스를 생성합니다
func NewCommandDependencies(
	storage interfaces.StorageRepository,
	apiClient interfaces.APIClient,
	scoreboardManager *ScoreboardManager,
	tierManager *models.TierManager,
	scoreCalculator interfaces.ScoreCalculator,
) *CommandDependencies {
	return &CommandDependencies{
		Storage:           storage,
		APIClient:         apiClient,
		ScoreboardManager: scoreboardManager,
		TierManager:       tierManager,
		ScoreCalculator:   scoreCalculator,
	}
}