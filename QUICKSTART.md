# Quick Start Guide

## 1. Get Your Discord Bot Token

1. Go to https://discord.com/developers/applications
2. Create a new application
3. Go to "Bot" section
4. Click "Reset Token" and copy your token
5. Enable "Message Content Intent" under "Privileged Gateway Intents"

## 2. Generate Configuration

```bash
make build
make generate
```

This creates a `config.yaml` file with examples.

## 3. Set Your Token

```bash
export DISCORD_BOT_TOKEN="your-token-here"
```

Or edit `config.yaml` and update the token field (not recommended for production).

## 4. Validate Configuration

```bash
make validate
```

## 5. Run the Bot

```bash
make run
```

Or with debug logging:

```bash
make run-debug
```

## 6. Test It

In Discord, type:
- `!ping` - Should respond with "Pong!"
- `!help` - Should show a help embed

## Common Commands

```bash
make help          # Show all available commands
make build         # Build the binary
make test          # Run tests
make docker-build  # Build Docker image
make docker-run    # Run in Docker
make clean         # Clean build artifacts
```

## Customizing Actions

Edit `config.yaml` to add your own actions. See README.md for full documentation.

Example - Add a custom command:

```yaml
actions:
  - name: "hello"
    description: "Says hello"
    type: "command"
    trigger:
      command: "hello"
    response:
      type: "text"
      content: "Hello from GXF Discord Bot!"
```

Save and restart the bot to apply changes.

## Using Vault for Secrets

1. Set up Vault/OpenBao
2. Store your token: `vault kv put secret/discord/bot token=YOUR_TOKEN`
3. Update `config.yaml`:

```yaml
bot:
  tokenVaultPath: "secret/discord/bot"

secrets:
  provider: "vault"
  address: "https://vault.example.com:8200"
  authMethod: "token"
  tokenEnvVar: "VAULT_TOKEN"
```

4. Set Vault token: `export VAULT_TOKEN="your-vault-token"`
5. Run the bot

## Deploying to Kubernetes

See README.md for complete Kubernetes deployment examples with ConfigMaps and Secrets.

## Troubleshooting

### Bot doesn't respond
- Check that "Message Content Intent" is enabled in Discord Developer Portal
- Verify bot has permissions in the Discord server
- Check logs for errors

### Configuration errors
- Run `make validate` to check for syntax errors
- Ensure CUE schema validation passes

### Connection issues
- Verify your bot token is correct
- Check network connectivity
- Review logs with `--debug` flag

## Next Steps

- Read README.md for complete documentation
- Explore example actions in the generated config.yaml
- Set up OAuth authentication for admin commands
- Configure scheduled tasks with cron syntax
- Integrate with external APIs using HTTP actions
