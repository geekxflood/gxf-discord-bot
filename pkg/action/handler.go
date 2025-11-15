// Package action provides action handling for Discord bot events.
package action

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
)

// Manager manages all bot actions
type Manager struct {
	actions []Action
	cfg     *config.Config
	logger  logging.Logger
}

// Action represents a bot action
type Action struct {
	Config  config.ActionConfig
	Handler Handler
}

// Handler is an interface for action handlers
type Handler interface {
	Matches(content string) bool
	Execute(ctx context.Context, session *discordgo.Session, message *discordgo.Message) error
}

// CommandHandler handles command-based actions
type CommandHandler struct {
	prefix  string
	command string
}

// MessageHandler handles pattern-based message actions
type MessageHandler struct {
	pattern *regexp.Regexp
}

// ReactionHandler handles reaction-based actions
type ReactionHandler struct {
	emoji string
}

// NewManager creates a new action manager
func NewManager(cfg *config.Config, logger logging.Logger) (*Manager, error) {
	logger.Info("Initializing action manager", "actionCount", len(cfg.Actions))

	mgr := &Manager{
		actions: make([]Action, 0),
		cfg:     cfg,
		logger:  logger,
	}

	// Initialize actions
	for _, actionCfg := range cfg.Actions {
		var handler Handler
		var err error

		switch actionCfg.Type {
		case "command":
			handler = NewCommandHandler(cfg.Bot.Prefix, actionCfg.Trigger.Command)
		case "message":
			handler, err = NewMessageHandler(actionCfg.Trigger.Pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to create message handler for %s: %w", actionCfg.Name, err)
			}
		case "reaction":
			handler = NewReactionHandler(actionCfg.Trigger.Emoji)
		default:
			logger.Debug("Unsupported action type", "type", actionCfg.Type, "name", actionCfg.Name)
			continue
		}

		mgr.actions = append(mgr.actions, Action{
			Config:  actionCfg,
			Handler: handler,
		})
	}

	logger.Info("Action manager initialized", "loadedActions", len(mgr.actions))
	return mgr, nil
}

// HandleMessage handles incoming messages
func (m *Manager) HandleMessage(ctx context.Context, session *discordgo.Session, message *discordgo.MessageCreate) error {
	for _, action := range m.actions {
		if action.Handler.Matches(message.Content) {
			m.logger.Debug("Action matched", "action", action.Config.Name, "content", message.Content)
			// TODO: Execute action handler
			return nil
		}
	}
	return nil
}

// HandleReaction handles reaction events
func (m *Manager) HandleReaction(ctx context.Context, session *discordgo.Session, reaction *discordgo.MessageReactionAdd) error {
	emojiName := reaction.Emoji.Name
	for _, action := range m.actions {
		if action.Config.Type == "reaction" && action.Handler.Matches(emojiName) {
			m.logger.Debug("Reaction action matched", "action", action.Config.Name, "emoji", emojiName)
			// TODO: Execute action handler
			return nil
		}
	}
	return nil
}

// GetActions returns all registered actions
func (m *Manager) GetActions() []config.ActionConfig {
	actions := make([]config.ActionConfig, len(m.actions))
	for i, action := range m.actions {
		actions[i] = action.Config
	}
	return actions
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(prefix, command string) *CommandHandler {
	return &CommandHandler{
		prefix:  prefix,
		command: strings.ToLower(command),
	}
}

// Matches checks if the content matches the command
func (h *CommandHandler) Matches(content string) bool {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, h.prefix) {
		return false
	}

	// Remove prefix
	content = strings.TrimPrefix(content, h.prefix)
	content = strings.TrimSpace(content)

	// Extract command (first word)
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return false
	}

	cmd := strings.ToLower(parts[0])
	return cmd == h.command
}

// ExtractArgs extracts arguments from the command
func (h *CommandHandler) ExtractArgs(content string) []string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, h.prefix)
	content = strings.TrimSpace(content)

	parts := strings.Fields(content)
	if len(parts) <= 1 {
		return []string{}
	}

	return parts[1:]
}

// Execute executes the command handler
func (h *CommandHandler) Execute(ctx context.Context, session *discordgo.Session, message *discordgo.Message) error {
	// TODO: Implement command execution
	return nil
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(pattern string) (*MessageHandler, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return &MessageHandler{
		pattern: regex,
	}, nil
}

// Matches checks if the content matches the pattern
func (h *MessageHandler) Matches(content string) bool {
	return h.pattern.MatchString(content)
}

// Execute executes the message handler
func (h *MessageHandler) Execute(ctx context.Context, session *discordgo.Session, message *discordgo.Message) error {
	// TODO: Implement message execution
	return nil
}

// NewReactionHandler creates a new reaction handler
func NewReactionHandler(emoji string) *ReactionHandler {
	return &ReactionHandler{
		emoji: emoji,
	}
}

// Matches checks if the reaction matches the emoji
func (h *ReactionHandler) Matches(reaction string) bool {
	return h.emoji == reaction
}

// Execute executes the reaction handler
func (h *ReactionHandler) Execute(ctx context.Context, session *discordgo.Session, message *discordgo.Message) error {
	// TODO: Implement reaction execution
	return nil
}
