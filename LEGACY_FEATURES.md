# Legacy Features Documentation

This document captures the features from the original implementation before the TDD rebuild.

## Core Features (to be reimplemented with TDD)

### 1. Configuration Management
- YAML-based configuration with CUE schema validation
- Support for multiple token sources (Vault, Env Var, Direct)
- Command prefix and bot status configuration

### 2. Secret Management (Vault/OpenBao)
- Multiple auth methods: token, approle, kubernetes
- Secure token retrieval from Vault/OpenBao
- TLS verification support

### 3. OAuth Authentication
- Support for Discord, Google, GitHub, custom providers
- User and role-based authorization
- Session management with duration control
- Callback server for OAuth flow

### 4. Action Types
- **command**: Prefix-based commands
- **message**: Pattern matching with regex
- **reaction**: Reaction event handling
- **scheduled**: Cron-based scheduled tasks

### 5. Response Types
- **text**: Plain text messages
- **embed**: Rich embeds with fields, colors, footers
- **dm**: Direct messages
- **reaction**: Add reactions
- **http**: HTTP webhooks
- **webhook**: Discord webhooks

### 6. Advanced Features
- Rate limiting (per-user, per-channel, per-guild, global)
- Conditions (role, user, channel, permission checks)
- Worker pool for concurrent action execution (10 workers, 100 task queue)
- Graceful shutdown with signal handling
- Health checks for Kubernetes

### 7. CLI Commands
- `generate`: Generate sample configuration
- `validate`: Validate configuration against CUE schema
- `run`: Start the bot (default command)

## Dependencies to Keep
- github.com/bwmarrin/discordgo - Discord API
- github.com/hashicorp/vault/api - Vault integration
- github.com/spf13/cobra - CLI framework
- github.com/robfig/cron/v3 - Scheduled tasks
- github.com/alitto/pond/v2 - Worker pool
- golang.org/x/oauth2 - OAuth support
- cuelang.org/go - Schema validation
- gopkg.in/yaml.v3 - YAML parsing

## Test Coverage (before rebuild)
- internal/config: 60.7%
- internal/handlers: 6.5%
- All other packages: 0%

## Goal: 100% Test Coverage with TDD
