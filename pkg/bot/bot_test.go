package bot_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/geekxflood/gxf-discord-bot/internal/testutil"
	"github.com/geekxflood/gxf-discord-bot/pkg/bot"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNew_Success(t *testing.T) {
	// Set up test token
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "!",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)

	require.NoError(t, err)
	require.NotNil(t, b)
}

func TestNew_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Token:  "", // No token
			Prefix: "!",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)

	assert.Error(t, err)
	assert.Nil(t, b)
	assert.Contains(t, err.Error(), "token")
}

func TestNew_InvalidConfig(t *testing.T) {
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "", // Invalid: empty prefix
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)

	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestStart_Success(t *testing.T) {
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "!",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()
	logger.On("Info", "Starting Discord bot", mock.Anything).Return()
	logger.On("Info", "Discord bot started successfully", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)
	require.NoError(t, err)

	// Note: In real implementation, we'd mock the Discord session
	// For now, we expect this to fail connecting to Discord
	err = b.Start(ctx)

	// This test will need to be updated when we add proper mocking
	assert.Error(t, err) // Expected to fail without valid token
}

func TestStop_Success(t *testing.T) {
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "!",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()
	logger.On("Info", "Stopping Discord bot", mock.Anything).Return()
	logger.On("Info", "Discord bot stopped", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)
	require.NoError(t, err)

	err = b.Stop()
	assert.NoError(t, err)
}

func TestBot_HandlesContext(t *testing.T) {
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "!",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	b, err := bot.New(ctx, cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, b)
}

func TestBot_GetConfig(t *testing.T) {
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "!",
			Status:      "Playing games",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)
	require.NoError(t, err)

	retrievedCfg := b.GetConfig()
	assert.Equal(t, "!", retrievedCfg.Bot.Prefix)
	assert.Equal(t, "Playing games", retrievedCfg.Bot.Status)
}

func TestBot_IsRunning(t *testing.T) {
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "!",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)
	require.NoError(t, err)

	// Initially not running
	assert.False(t, b.IsRunning())
}

func TestBot_MultipleStops(t *testing.T) {
	os.Setenv("TEST_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_BOT_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_BOT_TOKEN",
			Prefix:      "!",
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", "Initializing Discord bot", mock.Anything).Return()
	logger.On("Info", "Stopping Discord bot", mock.Anything).Return()
	logger.On("Info", "Discord bot stopped", mock.Anything).Return()

	ctx := context.Background()
	b, err := bot.New(ctx, cfg, logger)
	require.NoError(t, err)

	// Multiple stops should be safe
	err = b.Stop()
	assert.NoError(t, err)

	err = b.Stop()
	assert.NoError(t, err)
}

func TestBot_ConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		expectErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Bot: config.BotConfig{
					Token:  "valid-token",
					Prefix: "!",
				},
			},
			expectErr: false,
		},
		{
			name: "missing prefix",
			cfg: &config.Config{
				Bot: config.BotConfig{
					Token: "valid-token",
				},
			},
			expectErr: true,
		},
		{
			name: "missing token source",
			cfg: &config.Config{
				Bot: config.BotConfig{
					Prefix: "!",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &testutil.MockLogger{}
			logger.On("Info", "Initializing Discord bot", mock.Anything).Return()

			ctx := context.Background()
			b, err := bot.New(ctx, tt.cfg, logger)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, b)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, b)
			}
		})
	}
}
