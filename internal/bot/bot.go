// Package bot provides the core Discord bot functionality.
package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/internal/auth"
	"github.com/geekxflood/gxf-discord-bot/internal/config"
	"github.com/geekxflood/gxf-discord-bot/internal/handlers"
	"github.com/geekxflood/gxf-discord-bot/internal/secrets"
)

// Bot represents the Discord bot instance
type Bot struct {
	session     *discordgo.Session
	cfg         config.Provider
	secretsMgr  *secrets.Manager
	authMgr     *auth.Manager
	actionMgr   *handlers.ActionManager
	logger      logging.Logger
}

// New creates a new Discord bot instance
func New(ctx context.Context, cfg config.Provider, logger logging.Logger) (*Bot, error) {
	logger.Info("Initializing Discord bot")

	// Initialize secrets manager
	secretsMgr, err := secrets.NewManager(ctx, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	// Get bot token
	token, err := secretsMgr.GetBotToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot token: %w", err)
	}

	// Create Discord session
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Initialize auth manager
	authMgr, err := auth.NewManager(ctx, cfg, secretsMgr, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	bot := &Bot{
		session:    session,
		cfg:        cfg,
		secretsMgr: secretsMgr,
		authMgr:    authMgr,
		logger:     logger,
	}

	// Initialize action manager
	actionMgr, err := handlers.NewActionManager(cfg, authMgr, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create action manager: %w", err)
	}
	bot.actionMgr = actionMgr

	// Register event handlers
	bot.registerHandlers()

	// Set bot intents
	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

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

	// Set bot status
	if b.cfg.Exists("bot.status") {
		status, _ := b.cfg.GetString("bot.status", "")
		activityType, _ := b.cfg.GetString("bot.activityType", "playing")

		var at discordgo.ActivityType
		switch activityType {
		case "playing":
			at = discordgo.ActivityTypeGame
		case "streaming":
			at = discordgo.ActivityTypeStreaming
		case "listening":
			at = discordgo.ActivityTypeListening
		case "watching":
			at = discordgo.ActivityTypeWatching
		case "competing":
			at = discordgo.ActivityTypeCompeting
		default:
			at = discordgo.ActivityTypeGame
		}

		err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{
				{
					Name: status,
					Type: at,
				},
			},
			Status: "online",
		})

		if err != nil {
			b.logger.Error("Failed to set bot status", "error", err)
		}
	}
}

// handleMessageCreate handles message creation events
func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	ctx := context.Background()
	b.actionMgr.HandleMessage(ctx, s, m)
}

// handleMessageReactionAdd handles reaction add events
func (b *Bot) handleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore reactions from bots
	if r.Member != nil && r.Member.User.Bot {
		return
	}

	ctx := context.Background()
	b.actionMgr.HandleReaction(ctx, s, r)
}

// Start starts the Discord bot
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting Discord bot connection")

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	// Start scheduled actions
	b.actionMgr.StartScheduler(ctx, b.session)

	b.logger.Info("Discord bot started successfully")
	return nil
}

// Stop stops the Discord bot
func (b *Bot) Stop() error {
	b.logger.Info("Stopping Discord bot")

	// Stop scheduler
	b.actionMgr.StopScheduler()

	// Close auth manager
	if err := b.authMgr.Close(); err != nil {
		b.logger.Error("Error closing auth manager", "error", err)
	}

	// Close secrets manager
	if err := b.secretsMgr.Close(); err != nil {
		b.logger.Error("Error closing secrets manager", "error", err)
	}

	// Close Discord session
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("error closing Discord session: %w", err)
	}

	b.logger.Info("Discord bot stopped")
	return nil
}
