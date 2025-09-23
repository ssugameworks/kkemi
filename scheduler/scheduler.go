package scheduler

import (
	"sync"
	"time"

	"github.com/ssugameworks/Discord-Bot/bot"
	"github.com/ssugameworks/Discord-Bot/config"
	"github.com/ssugameworks/Discord-Bot/constants"
	"github.com/ssugameworks/Discord-Bot/sheets"
	"github.com/ssugameworks/Discord-Bot/utils"

	"github.com/bwmarrin/discordgo"
)

type Scheduler struct {
	session           *discordgo.Session
	config            *config.Config
	scoreboardManager *bot.ScoreboardManager
	sheetsClient      *sheets.SheetsClient
	ticker            *time.Ticker
	customTicker      *time.Ticker
	sheetsTicker      *time.Ticker
	stopChan          chan bool
	customStopChan    chan bool
	sheetsStopChan    chan bool
	mu                sync.Mutex
	stopped           bool
}

func NewScheduler(session *discordgo.Session, config *config.Config, scoreboardManager *bot.ScoreboardManager) *Scheduler {
	sheetsClient, err := sheets.NewSheetsClient()
	if err != nil {
		utils.Error("Failed to initialize sheets client: %v", err)
		sheetsClient = nil
	}

	return &Scheduler{
		session:           session,
		config:            config,
		scoreboardManager: scoreboardManager,
		sheetsClient:      sheetsClient,
		stopChan:          make(chan bool),
		customStopChan:    make(chan bool),
		sheetsStopChan:    make(chan bool),
	}
}

func (s *Scheduler) StartDailyScoreboard() {
	s.ticker = time.NewTicker(constants.SchedulerInterval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.sendDailyScoreboard()
			case <-s.stopChan:
				return
			}
		}
	}()

	utils.Info("Daily scoreboard scheduler started")
}

func (s *Scheduler) StartSheetsUpdate() {
	if s.sheetsClient == nil {
		utils.Warn("Sheets client not available - skipping sheets update scheduler")
		return
	}

	s.sheetsTicker = time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-s.sheetsTicker.C:
				s.updateSheetsScoreboard()
			case <-s.sheetsStopChan:
				return
			}
		}
	}()

	utils.Info("Sheets update scheduler started (30-minute interval)")
}

func (s *Scheduler) StartCustomSchedule(hour, minute int) {
	// 기존 커스텀 스케줄러가 있다면 정리
	s.stopCustomScheduler()

	now := utils.GetCurrentTimeKST()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	if nextRun.Before(now) {
		nextRun = nextRun.Add(constants.SchedulerInterval)
	}

	duration := nextRun.Sub(now)

	go func() {
		// 첫 실행까지 대기, 중단 신호 체크
		select {
		case <-time.After(duration):
			s.sendDailyScoreboard()
		case <-s.customStopChan:
			return
		}

		// 정기적 실행 시작
		s.customTicker = time.NewTicker(constants.SchedulerInterval)
		defer s.customTicker.Stop()

		for {
			select {
			case <-s.customTicker.C:
				s.sendDailyScoreboard()
			case <-s.customStopChan:
				return
			}
		}
	}()

	utils.Info("Daily scoreboard scheduler set to run daily at %02d:%02d", hour, minute)
}

func (s *Scheduler) sendDailyScoreboard() {
	if s.config.Discord.ChannelID == "" {
		utils.Error("Cannot send scoreboard: channel ID not configured")
		return
	}

	// 활성화된 대회가 있는지 확인
	storage := s.scoreboardManager.GetStorage()
	competition := storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		utils.Debug("No active competition - skipping daily scoreboard")
		return
	}

	// 대회 기간 내인지 확인
	now := utils.GetCurrentTimeKST()
	if now.Before(competition.StartDate) || now.After(competition.EndDate) {
		utils.Debug("Not within competition period - skipping daily scoreboard")
		return
	}

	// 블랙아웃 기간 확인 (마지막 날은 예외)
	isLastDay := now.Year() == competition.EndDate.Year() &&
		now.Month() == competition.EndDate.Month() &&
		now.Day() == competition.EndDate.Day()

	if storage.IsBlackoutPeriod() && !isLastDay {
		utils.Debug("Blackout period and not last day - skipping daily scoreboard")
		return
	}

	err := s.scoreboardManager.SendDailyScoreboard(s.session, s.config.Discord.ChannelID)
	if err != nil {
		utils.Error("Failed to send daily scoreboard: %v", err)
		return
	}

	utils.Info("Daily scoreboard sent successfully")
}

func (s *Scheduler) updateSheetsScoreboard() {
	if s.sheetsClient == nil {
		utils.Warn("Sheets client not available - skipping sheets update")
		return
	}

	// 활성화된 대회가 있는지 확인
	storage := s.scoreboardManager.GetStorage()
	competition := storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		utils.Debug("No active competition - skipping sheets update")
		return
	}

	// 대회 기간 내인지 확인
	now := utils.GetCurrentTimeKST()
	if now.Before(competition.StartDate) || now.After(competition.EndDate) {
		utils.Debug("Not within competition period - skipping sheets update")
		return
	}

	// 참가자 목록 가져오기
	participants := storage.GetParticipants()
	if len(participants) == 0 {
		utils.Debug("No participants found - skipping sheets update")
		return
	}

	// 점수 데이터 수집
	scores, err := s.scoreboardManager.CollectScoreData()
	if err != nil {
		utils.Error("Failed to collect score data for sheets: %v", err)
		return
	}

	// 스프레드시트 업데이트
	err = s.sheetsClient.UpdateScoreboardSheet("18w6gg5clIR5bUfOrObFUJQ8AMJquGUw6iBbPfVFZytE", scores)
	if err != nil {
		utils.Error("Failed to update sheets: %v", err)
		return
	}

	utils.Info("Successfully updated sheets scoreboard")
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return
	}
	s.stopped = true

	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}

	s.stopCustomSchedulerUnsafe()
	s.stopSheetsSchedulerUnsafe()

	// 채널 정리 - 논블로킹으로 신호 전송
	select {
	case s.stopChan <- true:
	default:
	}

	utils.Info("Scheduler stopped")
}

func (s *Scheduler) stopCustomScheduler() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopCustomSchedulerUnsafe()
}

func (s *Scheduler) stopCustomSchedulerUnsafe() {
	if s.customTicker != nil {
		s.customTicker.Stop()
		s.customTicker = nil
	}

	// 채널 정리 - 논블로킹으로 신호 전송
	select {
	case s.customStopChan <- true:
	default:
	}
}

func (s *Scheduler) stopSheetsSchedulerUnsafe() {
	if s.sheetsTicker != nil {
		s.sheetsTicker.Stop()
		s.sheetsTicker = nil
	}

	// 채널 정리 - 논블로킹으로 신호 전송
	select {
	case s.sheetsStopChan <- true:
	default:
	}
}
