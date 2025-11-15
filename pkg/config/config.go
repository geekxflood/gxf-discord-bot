// Package config provides configuration management for the Discord bot.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Bot     BotConfig      `yaml:"bot"`
	Actions []ActionConfig `yaml:"actions,omitempty"`
	Auth    *AuthConfig    `yaml:"auth,omitempty"`
	Secrets *SecretsConfig `yaml:"secrets,omitempty"`
}

// BotConfig contains Discord bot configuration
type BotConfig struct {
	Token         string `yaml:"token,omitempty"`
	TokenEnvVar   string `yaml:"tokenEnvVar,omitempty"`
	TokenVaultPath string `yaml:"tokenVaultPath,omitempty"`
	Prefix        string `yaml:"prefix"`
	Status        string `yaml:"status,omitempty"`
	ActivityType  string `yaml:"activityType,omitempty"`
}

// ActionConfig represents a bot action configuration
type ActionConfig struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Type        string         `yaml:"type"`
	Trigger     TriggerConfig  `yaml:"trigger"`
	Response    ResponseConfig `yaml:"response"`
	RequireAuth bool           `yaml:"requireAuth,omitempty"`
}

// TriggerConfig defines when an action is triggered
type TriggerConfig struct {
	Command  string   `yaml:"command,omitempty"`
	Pattern  string   `yaml:"pattern,omitempty"`
	Emoji    string   `yaml:"emoji,omitempty"`
	Schedule string   `yaml:"schedule,omitempty"`
	Channels []string `yaml:"channels,omitempty"`
}

// ResponseConfig defines how the bot responds
type ResponseConfig struct {
	Type     string       `yaml:"type"`
	Content  string       `yaml:"content,omitempty"`
	Embed    *EmbedConfig `yaml:"embed,omitempty"`
	Reaction string       `yaml:"reaction,omitempty"`
}

// EmbedConfig represents a Discord embed
type EmbedConfig struct {
	Title       string        `yaml:"title,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Color       int           `yaml:"color,omitempty"`
	Fields      []EmbedField  `yaml:"fields,omitempty"`
	Footer      string        `yaml:"footer,omitempty"`
	Timestamp   bool          `yaml:"timestamp,omitempty"`
}

// EmbedField represents a field in a Discord embed
type EmbedField struct {
	Name   string `yaml:"name"`
	Value  string `yaml:"value"`
	Inline bool   `yaml:"inline,omitempty"`
}

// AuthConfig contains OAuth authentication configuration
type AuthConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Provider        string   `yaml:"provider"`
	ClientID        string   `yaml:"clientId"`
	ClientSecretEnvVar string `yaml:"clientSecretEnvVar"`
	RedirectURL     string   `yaml:"redirectUrl"`
	Scopes          []string `yaml:"scopes,omitempty"`
	AuthorizedUsers []string `yaml:"authorizedUsers,omitempty"`
	AuthorizedRoles []string `yaml:"authorizedRoles,omitempty"`
}

// SecretsConfig contains secret management configuration
type SecretsConfig struct {
	Provider   string              `yaml:"provider"`
	Address    string              `yaml:"address"`
	AuthMethod string              `yaml:"authMethod"`
	MountPath  string              `yaml:"mountPath,omitempty"`
	TLSVerify  bool                `yaml:"tlsVerify,omitempty"`
	Kubernetes *KubernetesAuthConfig `yaml:"kubernetes,omitempty"`
	AppRole    *AppRoleAuthConfig  `yaml:"appRole,omitempty"`
	TokenEnvVar string              `yaml:"tokenEnvVar,omitempty"`
}

// KubernetesAuthConfig for Kubernetes authentication
type KubernetesAuthConfig struct {
	Role           string `yaml:"role"`
	ServiceAccount string `yaml:"serviceAccount"`
}

// AppRoleAuthConfig for AppRole authentication
type AppRoleAuthConfig struct {
	RoleID   string `yaml:"roleId"`
	SecretID string `yaml:"secretId"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// GetBotToken retrieves the bot token from configured sources
// Priority: Direct token > Environment variable > Vault
func (c *Config) GetBotToken() (string, error) {
	// Check direct token first
	if c.Bot.Token != "" {
		return c.Bot.Token, nil
	}

	// Check environment variable
	if c.Bot.TokenEnvVar != "" {
		token := os.Getenv(c.Bot.TokenEnvVar)
		if token == "" {
			return "", fmt.Errorf("environment variable %s not set", c.Bot.TokenEnvVar)
		}
		return token, nil
	}

	// Vault path would be handled by secrets manager
	if c.Bot.TokenVaultPath != "" {
		return "", fmt.Errorf("vault token retrieval requires secrets manager")
	}

	return "", fmt.Errorf("no token source configured")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate bot config
	if c.Bot.Prefix == "" {
		return fmt.Errorf("bot prefix is required")
	}

	// Ensure at least one token source is configured
	if c.Bot.Token == "" && c.Bot.TokenEnvVar == "" && c.Bot.TokenVaultPath == "" {
		return fmt.Errorf("no token source configured (token, tokenEnvVar, or tokenVaultPath required)")
	}

	return nil
}
