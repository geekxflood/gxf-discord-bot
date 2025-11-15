// Package bot provides the core Discord bot functionality.
package bot

import (
	"context"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
)

// Bot represents the Discord bot instance
type Bot struct {
	session  *discordgo.Session
	cfg      *config.Config
	logger   logging.Logger
	running  bool
	runningM sync.RWMutex
}

// New creates a new Discord bot instance
func New(ctx context.Context, cfg *config.Config, logger logging.Logger) (*Bot, error) {
	logger.Info("Initializing Discord bot")

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get bot token
	token, err := cfg.GetBotToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot token: %w", err)
	}

	// Create Discord session
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set bot intents
	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

	bot := &Bot{
		session: session,
		cfg:     cfg,
		logger:  logger,
		running: false,
	}

	// Register event handlers
	bot.registerHandlers()

	return bot, nil
}

// registerHandlers registers Discord event handlers
func (b *Bot) registerHandlers() {
	b.session.AddHandler(b.handleReady)
	b.session.AddHandler(b.handleMessageCreate)
	b.session.AddHandler(b.handleMessageReactionAdd)
}

// handleReady is called when the bot is ready
func (b *Bot) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	b.logger.Info("Bot is ready", "user", event.User.String(), "guilds", len(event.Guilds))

	// Set bot status if configured
	if b.cfg.Bot.Status != "" {
		activityType := b.getActivityType(b.cfg.Bot.ActivityType)

		err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{
				{
					Name: b.cfg.Bot.Status,
					Type: activityType,
				},
			},
			Status: "online",
		})

		if err != nil {
			b.logger.Error("Failed to set bot status", "error", err)
		}
	}
}

// getActivityType converts string to ActivityType
func (b *Bot) getActivityType(activityType string) discordgo.ActivityType {
	switch activityType {
	case "playing":
		return discordgo.ActivityTypeGame
	case "streaming":
		return discordgo.ActivityTypeStreaming
	case "listening":
		return discordgo.ActivityTypeListening
	case "watching":
		return discordgo.ActivityTypeWatching
	case "competing":
		return discordgo.ActivityTypeCompeting
	default:
		return discordgo.ActivityTypeGame
	}
}

// handleMessageCreate handles message creation events
func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	// TODO: Handle messages with action manager
	b.logger.Debug("Message received", "author", m.Author.Username, "content", m.Content)
}

// handleMessageReactionAdd handles reaction add events
func (b *Bot) handleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore reactions from bots
	if r.Member != nil && r.Member.User.Bot {
		return
	}

	// TODO: Handle reactions with action manager
	b.logger.Debug("Reaction added", "emoji", r.Emoji.Name)
}

// Start starts the Discord bot
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting Discord bot")

	b.runningM.Lock()
	defer b.runningM.Unlock()

	if b.running {
		return fmt.Errorf("bot is already running")
	}

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	b.running = true
	b.logger.Info("Discord bot started successfully")

	return nil
}

// Stop stops the Discord bot
func (b *Bot) Stop() error {
	b.runningM.Lock()
	defer b.runningM.Unlock()

	if !b.running && b.session == nil {
		// Already stopped or never started
		return nil
	}

	b.logger.Info("Stopping Discord bot")

	if b.session != nil {
		if err := b.session.Close(); err != nil {
			b.logger.Error("Error closing Discord session", "error", err)
			// Don't return error, continue cleanup
		}
	}

	b.running = false
	b.logger.Info("Discord bot stopped")

	return nil
}

// IsRunning returns whether the bot is currently running
func (b *Bot) IsRunning() bool {
	b.runningM.RLock()
	defer b.runningM.RUnlock()
	return b.running
}

// GetConfig returns the bot's configuration
func (b *Bot) GetConfig() *config.Config {
	return b.cfg
}
