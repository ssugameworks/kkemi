package scheduler

import (
	"discord-bot/bot"
	"discord-bot/config"
	"discord-bot/constants"
	"discord-bot/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Scheduler struct {
	session           *discordgo.Session
	config            *config.Config
	scoreboardManager *bot.ScoreboardManager
	ticker            *time.Ticker
	customTicker      *time.Ticker
	stopChan          chan bool
	customStopChan    chan bool
}

func NewScheduler(session *discordgo.Session, config *config.Config, scoreboardManager *bot.ScoreboardManager) *Scheduler {
	return &Scheduler{
		session:           session,
		config:            config,
		scoreboardManager: scoreboardManager,
		stopChan:          make(chan bool),
		customStopChan:    make(chan bool),
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

func (s *Scheduler) StartCustomSchedule(hour, minute int) {
	// 기존 커스텀 스케줄러가 있다면 정리
	s.stopCustomScheduler()

	now := time.Now()
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
	now := time.Now()
	if now.Before(competition.StartDate) || now.After(competition.EndDate) {
		utils.Debug("Not within competition period - skipping daily scoreboard")
		return
	}

	// 블랙아웃 기간 확인 (마지막 날은 예외)
	storage := s.scoreboardManager.GetStorage()
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

func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	s.stopCustomScheduler()

	select {
	case s.stopChan <- true:
	default:
		// channel is full or no receiver, skip
	}

	utils.Info("Scheduler stopped")
}

func (s *Scheduler) stopCustomScheduler() {
	if s.customTicker != nil {
		s.customTicker.Stop()
		s.customTicker = nil
	}

	select {
	case s.customStopChan <- true:
	default:
		// channel is full or no receiver, skip
	}
}
