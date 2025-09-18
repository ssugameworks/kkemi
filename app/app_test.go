package app

import (
	"os"
	"testing"

	"github.com/ssugameworks/Discord-Bot/config"
)

func TestApplication_New(t *testing.T) {
	// 환경변수 설정
	oldToken := os.Getenv("DISCORD_BOT_TOKEN")
	defer os.Setenv("DISCORD_BOT_TOKEN", oldToken)

	os.Setenv("DISCORD_BOT_TOKEN", "test-token-12345")

	t.Run("Successful application creation", func(t *testing.T) {
		app, err := New()

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if app == nil {
			t.Fatal("Expected non-nil application")
		}

		if app.config == nil {
			t.Error("Expected non-nil config")
		}

		if app.storage == nil {
			t.Error("Expected non-nil storage")
		}

		if app.apiClient == nil {
			t.Error("Expected non-nil API client")
		}

		if app.session == nil {
			t.Error("Expected non-nil Discord session")
		}

		if app.tierManager == nil {
			t.Error("Expected non-nil tier manager")
		}

		// Clean up
		if app.session != nil {
			app.Stop()
		}
	})

	t.Run("Missing token should fail", func(t *testing.T) {
		os.Setenv("DISCORD_BOT_TOKEN", "")

		app, err := New()

		if err == nil {
			t.Error("Expected error for missing token")
		}

		if app != nil {
			t.Error("Expected nil application on error")
			app.Stop()
		}
	})
}

func TestApplication_loadConfig(t *testing.T) {
	app := &Application{}

	// 유효한 토큰 설정
	oldToken := os.Getenv("DISCORD_BOT_TOKEN")
	defer os.Setenv("DISCORD_BOT_TOKEN", oldToken)

	os.Setenv("DISCORD_BOT_TOKEN", "test-token-12345")

	err := app.loadConfig()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if app.config == nil {
		t.Error("Expected non-nil config")
	}

	if app.config.Discord.Token != "test-token-12345" {
		t.Errorf("Expected token 'test-token-12345', got '%s'", app.config.Discord.Token)
	}
}

func TestApplication_initializeDependencies(t *testing.T) {
	app := &Application{
		config: &config.Config{
			Discord: config.DiscordConfig{
				Token: "test-token",
			},
		},
	}

	err := app.initializeDependencies()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if app.apiClient == nil {
		t.Error("Expected non-nil API client")
	}

	if app.storage == nil {
		t.Error("Expected non-nil storage")
	}
}

func TestApplication_initializeDiscord(t *testing.T) {
	app := &Application{
		config: &config.Config{
			Discord: config.DiscordConfig{
				Token: "test-token-12345",
			},
		},
	}

	err := app.initializeDiscord()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if app.session == nil {
		t.Error("Expected non-nil Discord session")
	}

	// 세션이 올바른 토큰으로 설정되었는지 확인
	if app.session.Token != "Bot test-token-12345" {
		t.Errorf("Expected token 'Bot test-token-12345', got '%s'", app.session.Token)
	}
}

func TestApplication_Stop(t *testing.T) {
	// 환경변수 설정
	oldToken := os.Getenv("DISCORD_BOT_TOKEN")
	defer os.Setenv("DISCORD_BOT_TOKEN", oldToken)

	os.Setenv("DISCORD_BOT_TOKEN", "test-token-12345")

	app, err := New()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Stop 메서드 테스트
	err = app.Stop()
	if err != nil {
		t.Errorf("Expected no error from Stop(), got: %v", err)
	}
}

func TestApplication_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Valid token",
			token:       "valid-token-12345",
			expectError: false,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "Short token",
			token:       "abc",
			expectError: false, // 토큰 길이 검증은 config에서 하지 않음
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 환경변수 설정
			oldToken := os.Getenv("DISCORD_BOT_TOKEN")
			defer os.Setenv("DISCORD_BOT_TOKEN", oldToken)

			os.Setenv("DISCORD_BOT_TOKEN", tt.token)

			app, err := New()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
					if app != nil {
						app.Stop()
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if app != nil {
					app.Stop()
				}
			}
		})
	}
}

func TestApplication_printStartupMessage(t *testing.T) {
	// 환경변수 설정
	oldToken := os.Getenv("DISCORD_BOT_TOKEN")
	defer os.Setenv("DISCORD_BOT_TOKEN", oldToken)

	os.Setenv("DISCORD_BOT_TOKEN", "test-token-12345")

	app, err := New()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}
	defer app.Stop()

	// printStartupMessage는 출력만 하므로 패닉이 발생하지 않는지 확인
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printStartupMessage panicked: %v", r)
		}
	}()

	app.printStartupMessage()
}

// 통합 테스트: 전체 애플리케이션 생성 및 종료
func TestApplication_Integration(t *testing.T) {
	// 환경변수 설정
	oldToken := os.Getenv("DISCORD_BOT_TOKEN")
	oldChannel := os.Getenv("DISCORD_CHANNEL_ID")
	defer func() {
		os.Setenv("DISCORD_BOT_TOKEN", oldToken)
		os.Setenv("DISCORD_CHANNEL_ID", oldChannel)
	}()

	os.Setenv("DISCORD_BOT_TOKEN", "test-token-12345")
	os.Setenv("DISCORD_CHANNEL_ID", "123456789")

	app, err := New()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// 모든 컴포넌트가 올바르게 초기화되었는지 확인
	if app.config == nil {
		t.Error("Config not initialized")
	}

	if app.storage == nil {
		t.Error("Storage not initialized")
	}

	if app.apiClient == nil {
		t.Error("API client not initialized")
	}

	if app.session == nil {
		t.Error("Discord session not initialized")
	}

	if app.tierManager == nil {
		t.Error("Tier manager not initialized")
	}

	if app.commandHandler == nil {
		t.Error("Command handler not initialized")
	}

	if app.scoreboardManager == nil {
		t.Error("Scoreboard manager not initialized")
	}

	if app.scheduler == nil {
		t.Error("Scheduler not initialized")
	}

	// 정상 종료
	err = app.Stop()
	if err != nil {
		t.Errorf("Failed to stop application: %v", err)
	}
}
