package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFile string
	force      bool

	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a sample configuration file",
		Long: `Generate a sample configuration file from the embedded CUE schema.
This creates a YAML file with example values and comments to help you get started.`,
		RunE: generateConfig,
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "config.yaml", "output file path")
	generateCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing file")
}

func generateConfig(cmd *cobra.Command, args []string) error {
	// Check if file exists
	if !force {
		if _, err := os.Stat(outputFile); err == nil {
			return fmt.Errorf("file %s already exists, use --force to overwrite", outputFile)
		}
	}

	sampleConfig := `# GXF Discord Bot Configuration
# This is a sample configuration file generated from the CUE schema

bot:
  # Discord bot token (use tokenEnvVar or tokenVaultPath in production)
  token: ""
  # Environment variable containing the token
  tokenEnvVar: "DISCORD_BOT_TOKEN"
  # Vault path to fetch token (e.g., "secret/data/discord/bot")
  # tokenVaultPath: "secret/data/discord/bot"
  # Command prefix
  prefix: "!"
  # Bot status message
  status: "Serving the community"
  # Activity type: playing, streaming, listening, watching, competing
  activityType: "playing"

# Secret store configuration (optional)
# Uncomment to enable Vault/OpenBao integration
# secrets:
#   provider: "vault"  # or "openbao"
#   address: "https://vault.example.com:8200"
#   authMethod: "token"  # token, approle, kubernetes
#   tokenEnvVar: "VAULT_TOKEN"
#   namespace: "secret"
#   mountPath: "secret"
#   tlsVerify: true

# Authentication configuration (optional)
# Uncomment to enable OAuth-based authentication
# auth:
#   enabled: true
#   provider: "discord"  # discord, google, github, custom
#   clientId: "your-oauth-client-id"
#   clientSecretEnvVar: "OAUTH_CLIENT_SECRET"
#   redirectUrl: "http://localhost:8080/callback"
#   scopes:
#     - "identify"
#     - "guilds"
#   authorizedUsers:
#     - "123456789012345678"  # Discord user IDs
#   sessionDuration: 60  # minutes
#   callbackServer:
#     host: "localhost"
#     port: 8080

# Logging configuration
logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"

# Bot actions
actions:
  # Simple ping command
  - name: "ping"
    description: "Responds with pong"
    type: "command"
    trigger:
      command: "ping"
    response:
      type: "text"
      content: "Pong! 🏓"

  # Help command with embed
  - name: "help"
    description: "Shows help information"
    type: "command"
    trigger:
      command: "help"
    response:
      type: "embed"
      embed:
        title: "Bot Commands"
        description: "Here are the available commands:"
        color: 3447003  # Blue
        fields:
          - name: "!ping"
            value: "Check if the bot is responsive"
            inline: false
          - name: "!help"
            value: "Show this help message"
            inline: false
        footer: "GXF Discord Bot"
        timestamp: true

  # Pattern-based message response
  - name: "greet"
    description: "Greets users who say hello"
    type: "message"
    trigger:
      pattern: "(?i)^(hello|hi|hey)\\s+(bot|everyone)"
    response:
      type: "text"
      content: "Hello! 👋 How can I help you today?"

  # Reaction-triggered action
  - name: "thumbs-up-response"
    description: "Responds to thumbs up reactions"
    type: "reaction"
    trigger:
      emoji: "👍"
    response:
      type: "dm"
      content: "Thanks for the thumbs up! 😊"

  # Admin command with conditions and rate limiting
  - name: "admin-info"
    description: "Admin-only command"
    type: "command"
    requireAuth: true
    trigger:
      command: "admin"
    response:
      type: "embed"
      embed:
        title: "Admin Panel"
        description: "Administrative information"
        color: 15158332  # Red
        fields:
          - name: "Status"
            value: "All systems operational"
            inline: true
        timestamp: true
    conditions:
      - type: "role"
        value: "ADMIN_ROLE_ID"  # Replace with actual role ID
    rateLimit:
      requests: 5
      window: 60
      scope: "user"

  # HTTP webhook example
  - name: "webhook-notify"
    description: "Sends notification via webhook"
    type: "command"
    trigger:
      command: "notify"
    response:
      type: "http"
      http:
        url: "https://your-webhook-endpoint.com/notify"
        method: "POST"
        headers:
          Content-Type: "application/json"
        body: '{"message": "Notification from Discord bot"}'
        timeout: 30

  # Scheduled task example
  - name: "daily-reminder"
    description: "Daily reminder at 9 AM"
    type: "scheduled"
    trigger:
      schedule: "0 9 * * *"  # Cron format: every day at 9 AM
      channels:
        - "CHANNEL_ID"  # Replace with actual channel ID
    response:
      type: "text"
      content: "Good morning! Don't forget to check your tasks for today! ☀️"
`

	if err := os.WriteFile(outputFile, []byte(sampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Configuration file generated: %s\n", outputFile)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the configuration file with your bot token and settings")
	fmt.Println("2. Validate the configuration: gxf-discord-bot validate")
	fmt.Println("3. Run the bot: gxf-discord-bot --config", outputFile)

	return nil
}
