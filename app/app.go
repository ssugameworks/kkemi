package app

import (
	"discord-bot/api"
	"discord-bot/bot"
	"discord-bot/config"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/scheduler"
	"discord-bot/scoring"
	"discord-bot/storage"
	"discord-bot/utils"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Application struct {
	config            *config.Config
	session           *discordgo.Session
	storage           interfaces.StorageRepository
	apiClient         interfaces.APIClient
	tierManager       *models.TierManager
	commandHandler    *bot.CommandHandler
	scoreboardManager *bot.ScoreboardManager
	scheduler         *scheduler.Scheduler
}

func New() (*Application, error) {
	app := &Application{}

	if err := app.loadConfig(); err != nil {
		return nil, err
	}

	if err := app.initializeDependencies(); err != nil {
		return nil, err
	}

	if err := app.initializeDiscord(); err != nil {
		return nil, err
	}

	app.setupHandlers()
	app.initializeScheduler()

	return app, nil
}

func (app *Application) loadConfig() error {
	app.config = config.Load()
	if err := app.config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}

func (app *Application) initializeDependencies() error {
	// 단일 API 클라이언트 인스턴스 생성
	app.apiClient = api.NewSolvedACClient()

	// API 클라이언트를 주입하여 Storage 생성
	app.storage = storage.NewStorage(app.apiClient)

	return nil
}

func (app *Application) initializeDiscord() error {
	session, err := discordgo.New("Bot " + app.config.Discord.Token)
	if err != nil {
		return fmt.Errorf("디스코드 세션 생성 실패: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuilds | discordgo.IntentsDirectMessages
	app.session = session
	return nil
}

func (app *Application) setupHandlers() {
	// 글로벌 TierManager 한 번만 생성
	app.tierManager = models.GetTierManager()
	
	// 의존성 주입을 통한 컴포넌트 생성
	calculator := scoring.NewScoreCalculator(app.apiClient, app.tierManager)
	app.scoreboardManager = bot.NewScoreboardManager(app.storage, calculator, app.apiClient, app.tierManager)
	app.commandHandler = bot.NewCommandHandler(app.storage, app.apiClient, app.scoreboardManager, app.tierManager)

	app.session.AddHandler(app.commandHandler.HandleMessage)
	app.session.AddHandler(app.handleReady)
}

func (app *Application) initializeScheduler() {
	app.scheduler = scheduler.NewScheduler(app.session, app.config, app.scoreboardManager)
}

func (app *Application) Start() error {
	if err := app.session.Open(); err != nil {
		return fmt.Errorf("웹소켓 연결 실패: %w", err)
	}

	if app.config.Schedule.Enabled {
		app.scheduler.StartCustomSchedule(
			app.config.Schedule.ScoreboardHour,
			app.config.Schedule.ScoreboardMinute,
		)
		utils.Info("매일 %02d:%02d에 자동으로 스코어보드가 띄워집니다.",
			app.config.Schedule.ScoreboardHour, app.config.Schedule.ScoreboardMinute)
	} else {
		log.Println("DISCORD_CHANNEL_ID가 설정되지 않았습니다. 스코어보드가 비활성화되었습니다.")
	}

	app.printStartupMessage()
	return nil
}

func (app *Application) printStartupMessage() {
	utils.Info("Discord Bot v0.1.0")
	fmt.Println("📋 사용 가능한 명령어: !help")
	if app.config.Schedule.Enabled {
		fmt.Printf("⏰ 매일 %02d:%02d에 자동으로 스코어보드가 전송됩니다.\n",
			app.config.Schedule.ScoreboardHour, app.config.Schedule.ScoreboardMinute)
	}
}

func (app *Application) Run() error {
	if err := app.Start(); err != nil {
		return err
	}

	// 종료 신호 대기
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGKILL)
	<-sc

	return app.Stop()
}

func (app *Application) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	// TODO: Welcome message
}

func (app *Application) Stop() error {
	fmt.Println("🔄 봇을 종료하는 중...")

	if app.scheduler != nil {
		app.scheduler.Stop()
	}

	if app.session != nil {
		app.session.Close()
	}

	fmt.Println("봇이 정상적으로 종료되었습니다.")
	return nil
}
