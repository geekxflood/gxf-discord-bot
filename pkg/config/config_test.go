package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
bot:
  token: "test-token-123"
  prefix: "!"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load the config
	cfg, err := config.Load(configPath)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "test-token-123", cfg.Bot.Token)
	assert.Equal(t, "!", cfg.Bot.Prefix)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := config.Load("nonexistent.yaml")

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `
bot:
  token: "test
    invalid yaml here
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	cfg, err := config.Load(configPath)

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestConfig_GetBotToken_FromDirectToken(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Token:  "direct-token",
			Prefix: "!",
		},
	}

	token, err := cfg.GetBotToken()

	require.NoError(t, err)
	assert.Equal(t, "direct-token", token)
}

func TestConfig_GetBotToken_FromEnvVar(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_DISCORD_TOKEN", "env-token-123")
	defer os.Unsetenv("TEST_DISCORD_TOKEN")

	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "TEST_DISCORD_TOKEN",
			Prefix:      "!",
		},
	}

	token, err := cfg.GetBotToken()

	require.NoError(t, err)
	assert.Equal(t, "env-token-123", token)
}

func TestConfig_GetBotToken_EnvVarNotSet(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			TokenEnvVar: "NONEXISTENT_VAR",
			Prefix:      "!",
		},
	}

	token, err := cfg.GetBotToken()

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "NONEXISTENT_VAR not set")
}

func TestConfig_GetBotToken_NoTokenProvided(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Prefix: "!",
		},
	}

	token, err := cfg.GetBotToken()

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "no token source configured")
}

func TestConfig_Validate_Success(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Token:  "valid-token",
			Prefix: "!",
		},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestConfig_Validate_MissingPrefix(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Token: "valid-token",
		},
	}

	err := cfg.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prefix is required")
}

func TestConfig_Validate_MissingToken(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Prefix: "!",
		},
	}

	err := cfg.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token source")
}
