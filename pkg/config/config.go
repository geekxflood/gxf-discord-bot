// Package config provides configuration management for the Discord bot.
package config

import (
	"fmt"
	"os"
	"regexp"

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
	// #nosec G304 -- Path is from command-line argument, expected behavior for config loading
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

	// Validate actions
	if err := c.validateActions(); err != nil {
		return fmt.Errorf("invalid action configuration: %w", err)
	}

	return nil
}

// validateActions validates all action configurations
func (c *Config) validateActions() error {
	actionNames := make(map[string]bool)

	for i, action := range c.Actions {
		// Check for duplicate names
		if action.Name == "" {
			return fmt.Errorf("action[%d]: name is required", i)
		}
		if actionNames[action.Name] {
			return fmt.Errorf("action[%d]: duplicate action name '%s'", i, action.Name)
		}
		actionNames[action.Name] = true

		// Validate action type
		validTypes := map[string]bool{
			"command":   true,
			"message":   true,
			"reaction":  true,
			"scheduled": true,
		}
		if !validTypes[action.Type] {
			return fmt.Errorf("action '%s': invalid type '%s' (must be command, message, reaction, or scheduled)", action.Name, action.Type)
		}

		// Validate trigger based on type
		if err := validateTrigger(action); err != nil {
			return fmt.Errorf("action '%s': %w", action.Name, err)
		}

		// Validate response
		if err := validateResponse(action); err != nil {
			return fmt.Errorf("action '%s': %w", action.Name, err)
		}
	}

	return nil
}

// validateTrigger validates the trigger configuration for an action
func validateTrigger(action ActionConfig) error {
	switch action.Type {
	case "command":
		if action.Trigger.Command == "" {
			return fmt.Errorf("command trigger requires 'command' field")
		}
	case "message":
		if action.Trigger.Pattern == "" {
			return fmt.Errorf("message trigger requires 'pattern' field")
		}
		// Validate regex pattern
		if _, err := regexp.Compile(action.Trigger.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", action.Trigger.Pattern, err)
		}
	case "reaction":
		if action.Trigger.Emoji == "" {
			return fmt.Errorf("reaction trigger requires 'emoji' field")
		}
	case "scheduled":
		if action.Trigger.Schedule == "" {
			return fmt.Errorf("scheduled trigger requires 'schedule' field")
		}
		if len(action.Trigger.Channels) == 0 {
			return fmt.Errorf("scheduled trigger requires at least one channel")
		}
	}
	return nil
}

// validateResponse validates the response configuration for an action
func validateResponse(action ActionConfig) error {
	validResponseTypes := map[string]bool{
		"text":     true,
		"embed":    true,
		"dm":       true,
		"reaction": true,
	}

	if !validResponseTypes[action.Response.Type] {
		return fmt.Errorf("invalid response type '%s' (must be text, embed, dm, or reaction)", action.Response.Type)
	}

	switch action.Response.Type {
	case "text":
		if action.Response.Content == "" {
			return fmt.Errorf("text response requires 'content' field")
		}
	case "embed":
		if action.Response.Embed == nil {
			return fmt.Errorf("embed response requires 'embed' configuration")
		}
	case "dm":
		if action.Response.Content == "" && action.Response.Embed == nil {
			return fmt.Errorf("dm response requires either 'content' or 'embed' field")
		}
	case "reaction":
		if action.Response.Reaction == "" {
			return fmt.Errorf("reaction response requires 'reaction' field")
		}
	}

	return nil
}
