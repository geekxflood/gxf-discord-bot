# GXF Discord Bot

A highly configurable Discord bot designed for Kubernetes deployments with enterprise-grade features including Vault/OpenBao secret management and OAuth-based authentication.

## Features

- **YAML-based Configuration** - Define bot behavior declaratively with CUE schema validation
- **Multiple Action Types** - Commands, pattern matching, reactions, scheduled tasks, HTTP webhooks
- **Secret Management** - Integrated support for HashiCorp Vault and OpenBao
- **OAuth Authentication** - Support for Discord, Google, GitHub, and custom OAuth providers
- **Rate Limiting** - Per-user, per-channel, per-guild, or global rate limits
- **Kubernetes Ready** - Designed for cloud-native deployments with health checks and graceful shutdown
- **Cobra CLI** - Generate, validate, and run commands for easy management

## Quick Start

### 1. Generate Configuration

```bash
gxf-discord-bot generate --output config.yaml
```

### 2. Configure Your Bot

Edit `config.yaml` and add your Discord bot token:

```yaml
bot:
  tokenEnvVar: "DISCORD_BOT_TOKEN"  # Recommended
  prefix: "!"
```

Set the environment variable:

```bash
export DISCORD_BOT_TOKEN="your-bot-token-here"
```

### 3. Validate Configuration

```bash
gxf-discord-bot validate --config config.yaml
```

### 4. Run the Bot

```bash
gxf-discord-bot --config config.yaml
```

## Configuration

### Bot Settings

```yaml
bot:
  # Token sources (priority: Vault > Env Var > Direct)
  tokenVaultPath: "secret/discord/bot"     # Vault path
  tokenEnvVar: "DISCORD_BOT_TOKEN"         # Environment variable
  token: "your-token"                       # Direct (not recommended)

  prefix: "!"                               # Command prefix
  status: "Serving the community"           # Bot status
  activityType: "playing"                   # playing, streaming, listening, watching
```

### Secret Store (Vault/OpenBao)

```yaml
secrets:
  provider: "vault"                         # or "openbao"
  address: "https://vault.example.com:8200"
  authMethod: "kubernetes"                  # token, approle, kubernetes

  # Kubernetes auth
  kubernetes:
    role: "discord-bot"
    serviceAccount: "/var/run/secrets/kubernetes.io/serviceaccount/token"

  # AppRole auth
  appRole:
    roleId: "your-role-id"
    secretId: "your-secret-id"

  # Token auth
  tokenEnvVar: "VAULT_TOKEN"

  mountPath: "secret"
  tlsVerify: true
```

### OAuth Authentication

```yaml
auth:
  enabled: true
  provider: "discord"                       # discord, google, github, custom
  clientId: "your-client-id"
  clientSecretEnvVar: "OAUTH_CLIENT_SECRET"
  redirectUrl: "http://localhost:8080/callback"
  scopes:
    - "identify"
    - "guilds"

  authorizedUsers:
    - "123456789012345678"                  # Discord user IDs

  authorizedRoles:
    - "987654321098765432"                  # Discord role IDs

  sessionDuration: 60                       # minutes

  callbackServer:
    host: "localhost"
    port: 8080
```

### Actions

#### Simple Command

```yaml
actions:
  - name: "ping"
    description: "Responds with pong"
    type: "command"
    trigger:
      command: "ping"
    response:
      type: "text"
      content: "Pong!"
```

#### Embed Response

```yaml
actions:
  - name: "help"
    description: "Shows help"
    type: "command"
    trigger:
      command: "help"
    response:
      type: "embed"
      embed:
        title: "Bot Commands"
        description: "Available commands"
        color: 3447003
        fields:
          - name: "!ping"
            value: "Check bot status"
        footer: "GXF Discord Bot"
        timestamp: true
```

#### Pattern Matching

```yaml
actions:
  - name: "greet"
    type: "message"
    trigger:
      pattern: "(?i)^(hello|hi)\\s+bot"
    response:
      type: "text"
      content: "Hello! How can I help?"
```

#### Reaction Handler

```yaml
actions:
  - name: "like-response"
    type: "reaction"
    trigger:
      emoji: "👍"
    response:
      type: "dm"
      content: "Thanks for the like!"
```

#### Scheduled Task

```yaml
actions:
  - name: "daily-reminder"
    type: "scheduled"
    trigger:
      schedule: "0 9 * * *"                 # Cron: 9 AM daily
      channels:
        - "CHANNEL_ID"
    response:
      type: "text"
      content: "Good morning!"
```

#### HTTP Webhook

```yaml
actions:
  - name: "notify"
    type: "command"
    trigger:
      command: "notify"
    response:
      type: "http"
      http:
        url: "https://api.example.com/notify"
        method: "POST"
        headers:
          Content-Type: "application/json"
        body: '{"message": "Notification from bot"}'
        timeout: 30
```

#### With Conditions and Rate Limiting

```yaml
actions:
  - name: "admin"
    type: "command"
    requireAuth: true
    trigger:
      command: "admin"
      channels:
        - "ADMIN_CHANNEL_ID"
    response:
      type: "text"
      content: "Admin command executed"
    conditions:
      - type: "role"
        value: "ADMIN_ROLE_ID"
    rateLimit:
      requests: 5
      window: 60                            # seconds
      scope: "user"                         # user, channel, guild, global
```

## Building

### Local Build

```bash
go build -o gxf-discord-bot
```

### Docker Build

```bash
docker build -t gxf-discord-bot:latest .
```

### Run with Docker

```bash
docker run -d \
  -e DISCORD_BOT_TOKEN="your-token" \
  -v $(pwd)/config.yaml:/app/config/config.yaml \
  gxf-discord-bot:latest
```

## Kubernetes Deployment

### Using Environment Variables

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: discord-bot-secret
type: Opaque
stringData:
  DISCORD_BOT_TOKEN: "your-token-here"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: discord-bot-config
data:
  config.yaml: |
    bot:
      tokenEnvVar: "DISCORD_BOT_TOKEN"
      prefix: "!"
    actions:
      - name: "ping"
        type: "command"
        trigger:
          command: "ping"
        response:
          type: "text"
          content: "Pong!"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: discord-bot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: discord-bot
  template:
    metadata:
      labels:
        app: discord-bot
    spec:
      containers:
      - name: bot
        image: gxf-discord-bot:latest
        envFrom:
        - secretRef:
            name: discord-bot-secret
        volumeMounts:
        - name: config
          mountPath: /app/config
      volumes:
      - name: config
        configMap:
          name: discord-bot-config
```

### Using Vault

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: discord-bot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: discord-bot
  template:
    metadata:
      labels:
        app: discord-bot
    spec:
      serviceAccountName: discord-bot
      containers:
      - name: bot
        image: gxf-discord-bot:latest
        volumeMounts:
        - name: config
          mountPath: /app/config
      volumes:
      - name: config
        configMap:
          name: discord-bot-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: discord-bot-config
data:
  config.yaml: |
    bot:
      tokenVaultPath: "secret/discord/bot"
      prefix: "!"

    secrets:
      provider: "vault"
      address: "https://vault.vault.svc.cluster.local:8200"
      authMethod: "kubernetes"
      kubernetes:
        role: "discord-bot"
      mountPath: "secret"

    actions:
      - name: "ping"
        type: "command"
        trigger:
          command: "ping"
        response:
          type: "text"
          content: "Pong!"
```

## CLI Commands

### Generate

Generate a sample configuration file:

```bash
gxf-discord-bot generate [flags]

Flags:
  -o, --output string   Output file path (default "config.yaml")
  -f, --force          Overwrite existing file
```

### Validate

Validate a configuration file:

```bash
gxf-discord-bot validate [flags]

Flags:
  --config string   Config file path (default "config.yaml")
```

### Run

Run the bot (default command):

```bash
gxf-discord-bot [flags]

Flags:
  --config string   Config file path (default "config.yaml")
  --debug          Enable debug logging
```

## Action Types

| Type | Description | Trigger | Response Types |
|------|-------------|---------|----------------|
| `command` | Prefix-based commands | Command name | text, embed, dm, http, webhook |
| `message` | Pattern matching | Regex pattern | text, embed, dm, http, webhook |
| `reaction` | Reaction events | Emoji | text, embed, dm |
| `scheduled` | Cron-based tasks | Cron schedule | text, embed, http, webhook |

## Response Types

| Type | Description | Configuration |
|------|-------------|---------------|
| `text` | Plain text message | `content` |
| `embed` | Rich embed | `embed` object |
| `dm` | Direct message | `content` or `embed` |
| `reaction` | Add reaction | `reaction` emoji |
| `http` | HTTP request | `http` object |
| `webhook` | Discord webhook | `webhookUrl` |

## Condition Types

| Type | Description | Value |
|------|-------------|-------|
| `role` | User has role | Role ID |
| `user` | Specific user | User ID |
| `channel` | Specific channel | Channel ID |
| `permission` | User has permission | Permission flag |

## Rate Limit Scopes

| Scope | Description |
|-------|-------------|
| `user` | Per user |
| `channel` | Per channel |
| `guild` | Per guild/server |
| `global` | Globally |

## Development

### Prerequisites

- Go 1.23+
- Discord Bot Token ([Get one here](https://discord.com/developers/applications))

### Install Dependencies

```bash
go mod download
```

### Run Tests

```bash
make test              # Run all tests
make test-race         # With race detector
make test-coverage     # Generate coverage report
make test-bench        # Run benchmarks
```

### Run Linter

```bash
make install-lint      # Install golangci-lint
make lint              # Run linter
make lint-fix          # Auto-fix issues
```

### Run Security Scan

```bash
make security          # Run gosec SAST
```

### Run All CI Checks Locally

```bash
make ci                # Lint + Test + Security
```

### Project Structure

```
gxf-discord-bot/
├── cmd/                    # CLI commands
│   ├── root.go            # Main command
│   ├── generate.go        # Generate command
│   ├── validate.go        # Validate command
│   └── schema/
│       └── config.cue     # CUE schema (embedded)
├── internal/
│   ├── auth/              # OAuth authentication
│   ├── bot/               # Discord bot core
│   ├── config/            # Configuration manager
│   ├── handlers/          # Action handlers
│   └── secrets/           # Vault/OpenBao integration
├── main.go                # Entry point
├── Dockerfile             # Container image
└── README.md             # This file
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DISCORD_BOT_TOKEN` | Discord bot token | Yes (if not in config/vault) |
| `VAULT_TOKEN` | Vault token | No (if using Vault with token auth) |
| `OAUTH_CLIENT_SECRET` | OAuth client secret | No (if using OAuth) |

## Logging

Configure logging in the config file:

```yaml
logging:
  level: "info"           # debug, info, warn, error
  format: "json"          # json, text
  output: "stdout"        # stdout, stderr, or file path
```

## Architecture

This project follows patterns from the `athena-backend` project:

- **Cobra CLI** - Command-line interface with generate, validate, and run commands
- **CUE Schema** - Embedded configuration schema for validation
- **Dependency Injection** - Config and logger injected into all services
- **Graceful Shutdown** - Proper signal handling and cleanup
- **Structured Logging** - JSON logging for cloud-native environments
- **Worker Pool** - Concurrent action execution using pond v2 (10 workers, 100 task queue)

### Performance Optimizations

- **Concurrent Processing** - Actions executed in parallel using worker pool
- **Rate Limiting** - Per-user/channel/guild/global limits with automatic cleanup
- **Caching** - OAuth sessions cached with expiration
- **Efficient Parsing** - Compiled regex patterns for message matching

## CI/CD

This project includes a comprehensive CI/CD pipeline with:

- **Automated Testing** - Unit tests with race detection and coverage
- **Linting** - golangci-lint with 25+ sub-linters
- **SAST** - Security scanning with Gosec and Trivy
- **Container Publishing** - Multi-platform Docker images to GHCR
- **Security Alerts** - SARIF results uploaded to GitHub Security tab

See [CI_CD.md](CI_CD.md) for detailed documentation.

### CI Badges

![CI Status](https://github.com/<owner>/gxf-discord-bot/workflows/CI/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## Support

For issues and questions, please open an issue on GitHub.
