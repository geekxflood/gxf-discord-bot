package schema

// Config defines the complete bot configuration schema
#Config: {
	// Bot configuration
	bot: #BotConfig

	// Secret store configuration (optional)
	secrets?: #SecretsConfig

	// Authentication configuration (optional)
	auth?: #AuthConfig

	// Actions define bot behaviors
	actions: [...#Action]

	// Logging configuration
	logging?: #LoggingConfig
}

// BotConfig contains Discord bot settings
#BotConfig: {
	// Discord bot token (can be from env, vault, or direct)
	token?: string

	// Environment variable containing the token
	tokenEnvVar?: string

	// Vault path to fetch token from (e.g., "secret/data/discord/bot")
	tokenVaultPath?: string

	// Command prefix (default: !)
	prefix: string | *"!"

	// Bot status message
	status?: string

	// Activity type: "playing", "streaming", "listening", "watching", "competing"
	activityType?: string | *"playing"
}

// SecretsConfig for Vault/OpenBao integration
#SecretsConfig: {
	// Provider: "vault" or "openbao"
	provider: "vault" | "openbao"

	// Vault/OpenBao address
	address: string

	// Authentication method: "token", "approle", "kubernetes"
	authMethod: string | *"token"

	// Token for token-based auth
	token?: string

	// Token from environment variable
	tokenEnvVar?: string

	// AppRole configuration
	appRole?: {
		roleId:   string
		secretId: string
	}

	// Kubernetes auth configuration
	kubernetes?: {
		role:            string
		serviceAccount?: string | *"/var/run/secrets/kubernetes.io/serviceaccount/token"
	}

	// Namespace for secrets (default: "secret")
	namespace?: string | *"secret"

	// Mount path (default: "secret")
	mountPath?: string | *"secret"

	// Enable TLS verification
	tlsVerify?: bool | *true

	// CA certificate path
	caCert?: string
}

// AuthConfig for SSO and authentication workflows
#AuthConfig: {
	// Enable authentication requirement
	enabled: bool | *false

	// OAuth2 provider: "discord", "google", "github", "custom"
	provider: "discord" | "google" | "github" | "custom"

	// OAuth2 client ID
	clientId: string

	// OAuth2 client secret (can be from vault)
	clientSecret?: string

	// Client secret from environment variable
	clientSecretEnvVar?: string

	// Client secret vault path
	clientSecretVaultPath?: string

	// OAuth2 redirect URL
	redirectUrl: string

	// OAuth2 scopes
	scopes: [...string]

	// Custom OAuth2 endpoints (for custom provider)
	endpoints?: {
		authUrl:  string
		tokenUrl: string
		userUrl:  string
	}

	// Authorized users (Discord user IDs)
	authorizedUsers?: [...string]

	// Authorized roles (Discord role IDs)
	authorizedRoles?: [...string]

	// Session duration in minutes
	sessionDuration?: int | *60

	// Web server for OAuth callback
	callbackServer?: {
		host: string | *"localhost"
		port: int & >=1 & <=65535 | *8080
	}
}

// LoggingConfig for structured logging
#LoggingConfig: {
	// Log level: "debug", "info", "warn", "error"
	level: "debug" | "info" | "warn" | "error" | *"info"

	// Log format: "json", "text"
	format: "json" | "text" | *"json"

	// Output: "stdout", "stderr", or file path
	output?: string | *"stdout"
}

// Action defines a bot action/command
#Action: {
	// Unique action name
	name: string

	// Human-readable description
	description: string

	// Action type: "command", "reaction", "message", "scheduled"
	type: "command" | "reaction" | "message" | "scheduled"

	// Trigger configuration
	trigger: #Trigger

	// Response configuration
	response: #Response

	// Execution conditions (optional)
	conditions?: [...#Condition]

	// Rate limiting (optional)
	rateLimit?: #RateLimit

	// Require authentication
	requireAuth?: bool | *false
}

// Trigger defines what triggers an action
#Trigger: {
	// Command name (for command type)
	command?: string

	// Regex pattern (for message type)
	pattern?: string

	// Emoji name or unicode (for reaction type)
	emoji?: string

	// Cron schedule (for scheduled type, e.g., "0 9 * * *")
	schedule?: string

	// Channel restrictions (Discord channel IDs)
	channels?: [...string]

	// Guild restrictions (Discord guild/server IDs)
	guilds?: [...string]
}

// Response defines how the bot responds
#Response: {
	// Response type: "text", "embed", "reaction", "dm", "http", "webhook"
	type: "text" | "embed" | "reaction" | "dm" | "http" | "webhook"

	// Text content
	content?: string

	// Embed configuration
	embed?: #Embed

	// Reaction emoji
	reaction?: string

	// HTTP request configuration
	http?: {
		url:     string
		method:  "GET" | "POST" | "PUT" | "DELETE" | *"POST"
		headers?: [string]: string
		body?:   string
		timeout?: int | *30
	}

	// Webhook URL
	webhookUrl?: string

	// Delete trigger message after response
	deleteAfter?: int

	// Ephemeral response (only visible to user)
	ephemeral?: bool | *false
}

// Embed for rich Discord embeds
#Embed: {
	title?:       string
	description?: string
	color?:       int
	fields?: [...#EmbedField]
	footer?:    string
	image?:     string
	thumbnail?: string
	author?: {
		name:    string
		iconUrl?: string
		url?:    string
	}
	timestamp?: bool | *false
}

// EmbedField represents a field in an embed
#EmbedField: {
	name:    string
	value:   string
	inline?: bool | *false
}

// Condition defines execution conditions
#Condition: {
	// Condition type: "role", "user", "channel", "permission", "time"
	type: "role" | "user" | "channel" | "permission" | "time"

	// Value depends on type
	value: string

	// Operator: "equals", "contains", "matches", "not"
	operator?: "equals" | "contains" | "matches" | "not" | *"equals"
}

// RateLimit defines rate limiting for actions
#RateLimit: {
	// Maximum requests
	requests: int & >0

	// Time window in seconds
	window: int & >0

	// Per user or global
	scope: "user" | "channel" | "guild" | "global" | *"user"
}
