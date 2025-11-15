// Package bot provides the core Discord bot functionality.
package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/pkg/action"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/geekxflood/gxf-discord-bot/pkg/ratelimit"
	"github.com/geekxflood/gxf-discord-bot/pkg/scheduler"
)

// Bot represents the Discord bot instance
type Bot struct {
	session    *discordgo.Session
	cfg        *config.Config
	logger     logging.Logger
	actionMgr  *action.Manager
	scheduler  *scheduler.Scheduler
	rateLimiter *ratelimit.Limiter
	running    bool
	runningM   sync.RWMutex
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

	// Initialize action manager
	actionMgr, err := action.NewManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create action manager: %w", err)
	}

	// Initialize optional scheduler
	sched := scheduler.New(logger)

	// Initialize optional rate limiter
	limiter := ratelimit.New(logger)

	bot := &Bot{
		session:     session,
		cfg:         cfg,
		logger:      logger,
		actionMgr:   actionMgr,
		scheduler:   sched,
		rateLimiter: limiter,
		running:     false,
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

	ctx := context.Background()
	if err := b.actionMgr.HandleMessage(ctx, s, m); err != nil {
		b.logger.Error("Failed to handle message", "error", err)
	}
}

// handleMessageReactionAdd handles reaction add events
func (b *Bot) handleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore reactions from bots
	if r.Member != nil && r.Member.User.Bot {
		return
	}

	ctx := context.Background()
	if err := b.actionMgr.HandleReaction(ctx, s, r); err != nil {
		b.logger.Error("Failed to handle reaction", "error", err)
	}
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

	// Start scheduler if configured
	if b.scheduler != nil {
		if err := b.scheduler.Start(); err != nil {
			b.logger.Error("Failed to start scheduler", "error", err)
		}
	}

	// Start rate limiter cleanup if configured
	if b.rateLimiter != nil {
		// Run cleanup every 5 minutes
		if err := b.rateLimiter.StartCleanup(5 * time.Minute); err != nil {
			b.logger.Error("Failed to start rate limiter cleanup", "error", err)
		}
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

	// Stop scheduler if running
	if b.scheduler != nil && b.scheduler.IsRunning() {
		if err := b.scheduler.Stop(); err != nil {
			b.logger.Error("Error stopping scheduler", "error", err)
		}
	}

	// Stop rate limiter cleanup
	if b.rateLimiter != nil {
		b.rateLimiter.StopCleanup()
	}

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

// GetScheduler returns the bot's scheduler
func (b *Bot) GetScheduler() *scheduler.Scheduler {
	return b.scheduler
}

// GetRateLimiter returns the bot's rate limiter
func (b *Bot) GetRateLimiter() *ratelimit.Limiter {
	return b.rateLimiter
}
