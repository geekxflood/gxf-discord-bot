# Full-Featured Bot Example

This example showcases all the features available in the GXF Discord Bot.

## Features Demonstrated

### 1. **Command Handling**
- `!ping` - Simple text response
- `!help` - Rich embed with multiple fields
- `!info` - Detailed bot information
- `!welcome` - DM response with embed

### 2. **Message Pattern Matching**
- Regex-based pattern matching
- Case-insensitive matching
- Auto-response to greetings and thanks

### 3. **Reaction Handling**
- Responds to emoji reactions (👍)
- Can trigger any response type

### 4. **Scheduled Tasks**
- Daily reminders using cron descriptors (`@daily`)
- Hourly status updates (`@hourly`)
- Supports full cron expressions

### 5. **Response Types**
- **Text**: Simple message responses
- **Embed**: Rich embedded messages with fields, colors, timestamps
- **DM**: Direct messages to users
- **Reaction**: Add emoji reactions to messages

## Running the Example

### Prerequisites

1. Create a Discord bot and get your token:
   - Go to https://discord.com/developers/applications
   - Create a New Application
   - Go to the Bot section and create a bot
   - Copy the bot token

2. Set up environment variable:
   ```bash
   export DISCORD_BOT_TOKEN="your-bot-token-here"
   ```

3. Invite the bot to your server:
   - Go to OAuth2 > URL Generator
   - Select scopes: `bot`, `applications.commands`
   - Select permissions: `Send Messages`, `Read Messages`, `Add Reactions`
   - Use the generated URL to invite the bot

### Running the Bot

```bash
# From the repository root
go run main.go --config examples/full-featured/config.yaml
```

Or build and run:

```bash
make build
./gxf-discord-bot --config examples/full-featured/config.yaml
```

## Customization

### Update Channel IDs

Replace the placeholder channel IDs in the scheduled tasks:

```yaml
trigger:
  schedule: "@daily"
  channels:
    - "YOUR_CHANNEL_ID_HERE"  # Right-click channel > Copy ID
```

### Modify Commands

Add your own commands by following the pattern:

```yaml
- name: "mycommand"
  description: "My custom command"
  type: "command"
  trigger:
    command: "mycommand"
  response:
    type: "text"
    content: "My response!"
```

### Add Scheduled Tasks

Use cron descriptors or full cron expressions:

```yaml
- name: "my-scheduled-task"
  type: "scheduled"
  trigger:
    schedule: "@weekly"  # or "0 0 * * 0" for Sunday midnight
    channels:
      - "CHANNEL_ID"
  response:
    type: "text"
    content: "Weekly reminder!"
```

Common cron descriptors:
- `@hourly` - Every hour
- `@daily` - Every day at midnight
- `@weekly` - Every Sunday at midnight
- `@monthly` - First day of month at midnight
- `@yearly` - January 1st at midnight

### Customize Embed Colors

Colors are in decimal format. Use a color picker and convert hex to decimal:

```yaml
color: 5814783  # Hex: #58B9FF (Blue)
color: 3447003  # Hex: #34A853 (Green)
color: 15844367 # Hex: #F1C40F (Gold)
color: 15158332 # Hex: #E74C3C (Red)
color: 9442302  # Hex: #9013FE (Purple)
```

## Rate Limiting

The bot includes built-in rate limiting (currently not configured in this example). To add rate limiting, see the main documentation.

## Next Steps

- Check out `/examples/basic/` for a simpler configuration
- Read the main README.md for full documentation
- Explore the `/pkg` directory for available packages
- Review test files for usage examples
