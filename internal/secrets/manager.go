// Package secrets provides secret management with Vault/OpenBao integration.
package secrets

import (
	"context"
	"fmt"
	"os"

	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/internal/config"
)

// Manager manages secret retrieval from various sources
type Manager struct {
	provider Provider
	cfg      config.Provider
	logger   logging.Logger
}

// NewManager creates a new secrets manager
func NewManager(ctx context.Context, cfg config.Provider, logger logging.Logger) (*Manager, error) {
	mgr := &Manager{
		cfg:    cfg,
		logger: logger,
	}

	// Initialize provider if secrets are configured
	if cfg.Exists("secrets") {
		provider, _ := cfg.GetString("secrets.provider", "vault")

		switch provider {
		case "vault", "openbao":
			vaultProvider, err := NewVaultProvider(ctx, cfg, logger)
			if err != nil {
				return nil, fmt.Errorf("failed to create vault provider: %w", err)
			}
			mgr.provider = vaultProvider
			logger.Info("Secrets manager initialized", "provider", provider)
		default:
			return nil, fmt.Errorf("unsupported secrets provider: %s", provider)
		}
	} else {
		logger.Info("No secrets provider configured, using environment variables")
	}

	return mgr, nil
}

// GetBotToken retrieves the Discord bot token from configured source
func (m *Manager) GetBotToken(ctx context.Context) (string, error) {
	// Priority: Vault > Environment Variable > Direct token

	// 1. Try Vault if configured
	if m.cfg.Exists("bot.tokenVaultPath") && m.provider != nil {
		path, _ := m.cfg.GetString("bot.tokenVaultPath", "")
		token, err := m.provider.GetSecretValue(ctx, path, "token")
		if err != nil {
			m.logger.Warn("Failed to get token from vault", "error", err)
		} else if token != "" {
			m.logger.Debug("Bot token retrieved from vault")
			return token, nil
		}
	}

	// 2. Try environment variable
	if m.cfg.Exists("bot.tokenEnvVar") {
		envVar, _ := m.cfg.GetString("bot.tokenEnvVar", "")
		token := os.Getenv(envVar)
		if token != "" {
			m.logger.Debug("Bot token retrieved from environment variable", "var", envVar)
			return token, nil
		}
	}

	// 3. Try direct token (not recommended)
	if m.cfg.Exists("bot.token") {
		token, _ := m.cfg.GetString("bot.token", "")
		if token != "" {
			m.logger.Warn("Using direct token from config (not recommended for production)")
			return token, nil
		}
	}

	return "", fmt.Errorf("bot token not found in any configured source")
}

// GetOAuthClientSecret retrieves OAuth client secret from configured source
func (m *Manager) GetOAuthClientSecret(ctx context.Context) (string, error) {
	if !m.cfg.Exists("auth") {
		return "", fmt.Errorf("auth configuration not found")
	}

	// Try Vault
	if m.cfg.Exists("auth.clientSecretVaultPath") && m.provider != nil {
		path, _ := m.cfg.GetString("auth.clientSecretVaultPath", "")
		secret, err := m.provider.GetSecretValue(ctx, path, "clientSecret")
		if err != nil {
			m.logger.Warn("Failed to get client secret from vault", "error", err)
		} else if secret != "" {
			m.logger.Debug("OAuth client secret retrieved from vault")
			return secret, nil
		}
	}

	// Try environment variable
	if m.cfg.Exists("auth.clientSecretEnvVar") {
		envVar, _ := m.cfg.GetString("auth.clientSecretEnvVar", "")
		secret := os.Getenv(envVar)
		if secret != "" {
			m.logger.Debug("OAuth client secret retrieved from environment variable", "var", envVar)
			return secret, nil
		}
	}

	// Try direct secret
	if m.cfg.Exists("auth.clientSecret") {
		secret, _ := m.cfg.GetString("auth.clientSecret", "")
		if secret != "" {
			m.logger.Warn("Using direct OAuth client secret from config (not recommended)")
			return secret, nil
		}
	}

	return "", fmt.Errorf("OAuth client secret not found in any configured source")
}

// GetSecret retrieves a secret from the configured provider
func (m *Manager) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	if m.provider == nil {
		return nil, fmt.Errorf("no secrets provider configured")
	}
	return m.provider.GetSecret(ctx, path)
}

// GetSecretValue retrieves a specific value from a secret
func (m *Manager) GetSecretValue(ctx context.Context, path, key string) (string, error) {
	if m.provider == nil {
		return "", fmt.Errorf("no secrets provider configured")
	}
	return m.provider.GetSecretValue(ctx, path, key)
}

// Close closes the secrets manager
func (m *Manager) Close() error {
	if m.provider != nil {
		return m.provider.Close()
	}
	return nil
}
