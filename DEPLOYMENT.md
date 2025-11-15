# Deployment Guide

This guide covers various deployment options for the GXF Discord Bot.

## Prerequisites

1. **Discord Bot Token**
   - Go to https://discord.com/developers/applications
   - Create a New Application
   - Navigate to the Bot section
   - Click "Add Bot" and confirm
   - Copy the bot token (you'll need this)

2. **Bot Permissions**
   - In the OAuth2 > URL Generator section:
     - Scopes: `bot`, `applications.commands`
     - Bot Permissions:
       - Send Messages
       - Read Messages/View Channels
       - Add Reactions
       - Embed Links
       - Read Message History
   - Use the generated URL to invite the bot to your server

## Local Development

### 1. Set Up Environment

```bash
# Clone the repository
git clone https://github.com/geekxflood/gxf-discord-bot
cd gxf-discord-bot

# Set your bot token
export DISCORD_BOT_TOKEN="your-bot-token-here"
```

### 2. Configure the Bot

Create or modify `config.yaml`:

```yaml
bot:
  tokenEnvVar: "DISCORD_BOT_TOKEN"
  prefix: "!"
  status: "Development Mode"
  activityType: "playing"

actions:
  - name: "ping"
    type: "command"
    trigger:
      command: "ping"
    response:
      type: "text"
      content: "Pong! 🏓"
```

### 3. Run the Bot

```bash
# Using go run
go run main.go --config config.yaml

# Or build and run
make build
./gxf-discord-bot --config config.yaml

# With debug logging
./gxf-discord-bot --config config.yaml --debug
```

## Docker Deployment

### 1. Create Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gxf-discord-bot .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/gxf-discord-bot .
COPY config.yaml .

CMD ["./gxf-discord-bot", "--config", "config.yaml"]
```

### 2. Build and Run

```bash
# Build the image
docker build -t gxf-discord-bot:latest .

# Run the container
docker run -d \
  --name discord-bot \
  -e DISCORD_BOT_TOKEN="your-token" \
  gxf-discord-bot:latest

# View logs
docker logs -f discord-bot

# Stop the bot
docker stop discord-bot
```

## Kubernetes Deployment

### 1. Create Secret

```bash
kubectl create secret generic discord-bot-token \
  --from-literal=token=your-bot-token-here
```

### 2. Create ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: discord-bot-config
data:
  config.yaml: |
    bot:
      tokenEnvVar: "DISCORD_BOT_TOKEN"
      prefix: "!"
      status: "Running on Kubernetes"
      activityType: "watching"

    actions:
      - name: "ping"
        type: "command"
        trigger:
          command: "ping"
        response:
          type: "text"
          content: "Pong! 🏓 (from Kubernetes)"
```

### 3. Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: discord-bot
  labels:
    app: discord-bot
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
      - name: discord-bot
        image: gxf-discord-bot:latest
        imagePullPolicy: IfNotPresent
        env:
        - name: DISCORD_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: discord-bot-token
              key: token
        volumeMounts:
        - name: config
          mountPath: /root/config.yaml
          subPath: config.yaml
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
      volumes:
      - name: config
        configMap:
          name: discord-bot-config
```

### 4. Deploy

```bash
# Apply the ConfigMap
kubectl apply -f configmap.yaml

# Apply the Deployment
kubectl apply -f deployment.yaml

# Check status
kubectl get pods -l app=discord-bot

# View logs
kubectl logs -f deployment/discord-bot

# Scale if needed
kubectl scale deployment/discord-bot --replicas=1
```

## Systemd Service (Linux)

### 1. Create Service File

Create `/etc/systemd/system/discord-bot.service`:

```ini
[Unit]
Description=GXF Discord Bot
After=network.target

[Service]
Type=simple
User=discord-bot
WorkingDirectory=/opt/discord-bot
Environment="DISCORD_BOT_TOKEN=your-token-here"
ExecStart=/opt/discord-bot/gxf-discord-bot --config /opt/discord-bot/config.yaml
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 2. Install and Start

```bash
# Create user
sudo useradd -r -s /bin/false discord-bot

# Create directory
sudo mkdir -p /opt/discord-bot
sudo chown discord-bot:discord-bot /opt/discord-bot

# Copy files
sudo cp gxf-discord-bot /opt/discord-bot/
sudo cp config.yaml /opt/discord-bot/
sudo chown discord-bot:discord-bot /opt/discord-bot/*

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable discord-bot
sudo systemctl start discord-bot

# Check status
sudo systemctl status discord-bot

# View logs
sudo journalctl -u discord-bot -f
```

## Production Best Practices

### 1. Security

- ✅ **Never commit tokens** - Use environment variables or secrets management
- ✅ **Use read-only filesystems** in containers where possible
- ✅ **Run as non-root user** in containers and systemd services
- ✅ **Enable TLS verification** for external services
- ✅ **Rotate tokens** regularly

### 2. Monitoring

```yaml
# Add health check endpoint (future enhancement)
# For now, monitor process and logs

# Kubernetes liveness probe
livenessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - pgrep -f gxf-discord-bot
  initialDelaySeconds: 30
  periodSeconds: 10
```

### 3. Resource Limits

Recommended resource limits:
- **Memory**: 128Mi-256Mi
- **CPU**: 100m-200m
- **Storage**: Minimal (< 100Mi for binary + config)

### 4. Logging

The bot uses structured logging. Configure log level:

```bash
# Debug mode (development)
./gxf-discord-bot --config config.yaml --debug

# Production (info level by default)
./gxf-discord-bot --config config.yaml
```

### 5. Rate Limiting Configuration

Configure rate limits in your config to prevent abuse:

```yaml
# Note: Rate limiter is initialized but not yet auto-configured
# Manual configuration through bot.GetRateLimiter() in custom code
```

### 6. Scheduled Tasks

For scheduled tasks, ensure:
- Only one bot instance is running (no replicas)
- Or use distributed locking for multiple instances
- Channel IDs are correctly configured

## Troubleshooting

### Bot Not Responding

1. Check bot is online in Discord server
2. Verify token is correct
3. Check bot has proper permissions
4. Review logs for errors

```bash
# Docker
docker logs discord-bot

# Kubernetes
kubectl logs -f deployment/discord-bot

# Systemd
sudo journalctl -u discord-bot -f
```

### Permission Errors

Ensure bot has these permissions in your server:
- View Channels
- Send Messages
- Embed Links
- Add Reactions
- Read Message History

### High Memory Usage

If memory usage is high:
1. Check for memory leaks in custom actions
2. Reduce rate limiter cleanup frequency
3. Limit scheduled tasks
4. Review message caching

## Scaling Considerations

### Horizontal Scaling

⚠️ **Important**: Due to scheduled tasks, running multiple replicas may cause duplicate executions. Options:

1. **Single Instance** (Recommended)
   ```bash
   kubectl scale deployment/discord-bot --replicas=1
   ```

2. **Leader Election** (Future Enhancement)
   - Implement leader election for scheduled tasks
   - Only leader executes scheduled tasks

3. **Separate Scheduler** (Advanced)
   - Run scheduler in separate deployment
   - Main bot handles only events

### Vertical Scaling

Adjust resources based on load:

```yaml
resources:
  requests:
    memory: "256Mi"  # Increase if needed
    cpu: "200m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## Backup and Recovery

### Configuration Backup

```bash
# Backup config
kubectl get configmap discord-bot-config -o yaml > backup-config.yaml

# Restore config
kubectl apply -f backup-config.yaml
```

### Token Rotation

```bash
# Update secret
kubectl create secret generic discord-bot-token \
  --from-literal=token=new-token \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart deployment
kubectl rollout restart deployment/discord-bot
```

## Support

For issues or questions:
- Review logs for error messages
- Check Discord Developer Portal for bot status
- Verify configuration is valid YAML
- Ensure all required permissions are granted
