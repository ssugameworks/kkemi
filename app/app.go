package app

import (
	"discord-bot/api"
	"discord-bot/bot"
	"discord-bot/config"
	"discord-bot/constants"
	"discord-bot/health"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/scheduler"
	"discord-bot/scoring"
	"discord-bot/storage"
	"discord-bot/utils"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/firestore"
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
	// 캐시된 API 클라이언트 인스턴스 생성
	app.apiClient = api.NewCachedSolvedACClient()

	// API 클라이언트를 주입하여 Storage 생성
	storage, err := storage.NewStorage(app.apiClient)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	app.storage = storage

	// Firestore 헬스체크 등록 (타입 확인을 위한 인터페이스 메서드 사용)
	type ClientProvider interface {
		GetClient() interface{}
	}
	
	if clientProvider, ok := storage.(ClientProvider); ok {
		if client := clientProvider.GetClient(); client != nil {
			if firestoreClient, ok := client.(*firestore.Client); ok && firestoreClient != nil {
				healthChecker := health.NewFirestoreHealthChecker(firestoreClient)
				health.RegisterHealthChecker("firestore", healthChecker)
				utils.Info("Firestore health checker registered")
			}
		}
	}

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
	app.commandHandler = bot.NewCommandHandler(app.storage, app.apiClient, app.scoreboardManager, app.tierManager, calculator)

	app.session.AddHandler(app.commandHandler.HandleMessage)
	app.session.AddHandler(app.handleReady)

	// 캐시 워밍업 - 기존 참가자 데이터로 캐시 미리 로드
	app.warmupCache()
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
		utils.Warn("DISCORD_CHANNEL_ID가 설정되지 않았습니다. 스코어보드가 비활성화되었습니다.")
	}

	app.printStartupMessage()
	return nil
}

func (app *Application) printStartupMessage() {
	utils.Info("Discord Bot v0.1.0")
	utils.Info("📋 사용 가능한 명령어: !help")
	if app.config.Schedule.Enabled {
		utils.Info("⏰ 매일 %02d:%02d에 자동으로 스코어보드가 전송됩니다.",
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
	utils.Info("Discord bot connected successfully as %s#%s", event.User.Username, event.User.Discriminator)
	utils.Info("Bot is serving %d guilds", len(event.Guilds))
	
	// 봇 상태 설정
	err := s.UpdateGameStatus(0, constants.BotStatusMessage)
	if err != nil {
		utils.Warn("Failed to set bot status: %v", err)
	}
}

// warmupCache 기존 참가자 데이터로 캐시를 미리 워밍업합니다
func (app *Application) warmupCache() {
	participants := app.storage.GetParticipants()
	if len(participants) == 0 {
		utils.Info("No participants found, skipping cache warmup")
		return
	}

	handles := make([]string, len(participants))
	for i, participant := range participants {
		handles[i] = participant.BaekjoonID
	}

	if cachedClient, ok := app.apiClient.(*api.CachedSolvedACClient); ok {
		cachedClient.WarmupCache(handles)
	}
}

// printCacheStats 캐시 통계를 출력합니다
func (app *Application) printCacheStats() {
	if cachedClient, ok := app.apiClient.(*api.CachedSolvedACClient); ok {
		stats := cachedClient.GetCacheStats()
		utils.Info("📊 %s", stats.String())
	}
}

func (app *Application) Stop() error {
	utils.Info("🔄 봇을 종료하는 중...")

	// 종료 전 캐시 통계 출력
	app.printCacheStats()

	if app.scheduler != nil {
		app.scheduler.Stop()
	}

	// API 클라이언트 종료
	if app.apiClient != nil {
		if cachedClient, ok := app.apiClient.(*api.CachedSolvedACClient); ok {
			cachedClient.Close()
		}
	}

	if app.session != nil {
		app.session.Close()
	}

	utils.Info("봇이 정상적으로 종료되었습니다.")
	return nil
}
