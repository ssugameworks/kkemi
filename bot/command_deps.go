package bot

import (
	"github.com/ssugameworks/Discord-Bot/constants"
	"github.com/ssugameworks/Discord-Bot/interfaces"
	"github.com/ssugameworks/Discord-Bot/models"
	"github.com/ssugameworks/Discord-Bot/sheets"
	"github.com/ssugameworks/Discord-Bot/telemetry"
	"github.com/ssugameworks/Discord-Bot/utils"

	"github.com/bwmarrin/discordgo"
)

// CommandDependencies 명령어 핸들러가 필요로 하는 모든 의존성을 묶어서 관리합니다
type CommandDependencies struct {
	Storage           interfaces.StorageRepository
	APIClient         interfaces.APIClient
	ScoreboardManager *ScoreboardManager
	TierManager       *models.TierManager
	ScoreCalculator   interfaces.ScoreCalculator
	Session           *discordgo.Session
	MetricsClient     *telemetry.MetricsClient
	SheetsClient      *sheets.SheetsClient
}

// NewCommandDependencies 새로운 CommandDependencies 인스턴스를 생성합니다
func NewCommandDependencies(
	storage interfaces.StorageRepository,
	apiClient interfaces.APIClient,
	scoreboardManager *ScoreboardManager,
	tierManager *models.TierManager,
	scoreCalculator interfaces.ScoreCalculator,
	session *discordgo.Session,
	metricsClient *telemetry.MetricsClient,
	sheetsClient *sheets.SheetsClient,
) *CommandDependencies {
	return &CommandDependencies{
		Storage:           storage,
		APIClient:         apiClient,
		ScoreboardManager: scoreboardManager,
		TierManager:       tierManager,
		ScoreCalculator:   scoreCalculator,
		Session:           session,
		MetricsClient:     metricsClient,
		SheetsClient:      sheetsClient,
	}
}

// UpdateBotStatus 봇 상태를 현재 대회에 맞게 업데이트합니다
func (deps *CommandDependencies) UpdateBotStatus() {
	if deps.Session == nil {
		return
	}

	statusMessage := constants.BotStatusMessage
	if competition := deps.Storage.GetCompetition(); competition != nil && competition.IsActive {
		statusMessage = competition.Name
	}

	err := deps.Session.UpdateGameStatus(0, statusMessage)
	if err != nil {
		utils.Warn("Failed to set bot status: %v", err)
	}
}
