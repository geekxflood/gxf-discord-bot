// Package response provides response handling for Discord bot actions.
package response

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
)

const (
	// MaxMessageLength is the maximum length of a Discord message
	MaxMessageLength = 2000
	// DefaultTimeout is the default timeout for Discord API calls
	DefaultTimeout = 10 * time.Second
)

// DiscordSession defines the interface for Discord session methods we need
type DiscordSession interface {
	ChannelMessageSend(channelID, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)
	UserChannelCreate(userID string, options ...discordgo.RequestOption) (*discordgo.Channel, error)
	MessageReactionAdd(channelID, messageID, emojiID string, options ...discordgo.RequestOption) error
}

// Execute executes a response based on the configuration
func Execute(ctx context.Context, session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig, logger logging.Logger) error {
	// Validate inputs
	if session == nil {
		return fmt.Errorf("session is nil")
	}
	if message == nil {
		return fmt.Errorf("message is nil")
	}
	if message.ChannelID == "" {
		return fmt.Errorf("message channel ID is empty")
	}

	// Add timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	logger.Debug("Executing response", "type", cfg.Type, "channelID", message.ChannelID)

	// Check context before executing
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled before execution: %w", ctx.Err())
	default:
	}

	switch cfg.Type {
	case "text":
		return executeTextResponse(ctx, session, message, cfg)
	case "embed":
		return executeEmbedResponse(ctx, session, message, cfg)
	case "dm":
		return executeDMResponse(ctx, session, message, cfg)
	case "reaction":
		return executeReactionResponse(ctx, session, message, cfg)
	default:
		return fmt.Errorf("unsupported response type: %s", cfg.Type)
	}
}

// executeTextResponse sends a text message to the channel
func executeTextResponse(ctx context.Context, session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	if cfg.Content == "" {
		return fmt.Errorf("text response requires non-empty content")
	}

	// Validate message length
	if len(cfg.Content) > MaxMessageLength {
		return fmt.Errorf("message content exceeds maximum length of %d characters (got %d)", MaxMessageLength, len(cfg.Content))
	}

	// Check context before API call
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	_, err := session.ChannelMessageSend(message.ChannelID, cfg.Content)
	if err != nil {
		return fmt.Errorf("failed to send text message: %w", err)
	}

	return nil
}

// executeEmbedResponse sends an embed message to the channel
func executeEmbedResponse(ctx context.Context, session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	if cfg.Embed == nil {
		return fmt.Errorf("embed response requires non-nil embed config")
	}

	// Check context before API call
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	embed := BuildEmbed(cfg.Embed)

	_, err := session.ChannelMessageSendEmbed(message.ChannelID, embed)
	if err != nil {
		return fmt.Errorf("failed to send embed: %w", err)
	}

	return nil
}

// executeDMResponse sends a direct message to the user
func executeDMResponse(ctx context.Context, session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	// Validate author exists
	if message.Author == nil {
		return fmt.Errorf("message author is nil, cannot send DM")
	}
	if message.Author.ID == "" {
		return fmt.Errorf("message author ID is empty, cannot send DM")
	}

	// Check context before API call
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Create DM channel
	channel, err := session.UserChannelCreate(message.Author.ID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel: %w", err)
	}

	// Validate content length if sending text
	if cfg.Content != "" && len(cfg.Content) > MaxMessageLength {
		return fmt.Errorf("DM content exceeds maximum length of %d characters (got %d)", MaxMessageLength, len(cfg.Content))
	}

	// Send message to DM channel
	content := cfg.Content
	if content == "" && cfg.Embed != nil {
		// If no content but embed exists, send embed
		embed := BuildEmbed(cfg.Embed)
		_, err = session.ChannelMessageSendEmbed(channel.ID, embed)
	} else {
		_, err = session.ChannelMessageSend(channel.ID, content)
	}

	if err != nil {
		return fmt.Errorf("failed to send DM: %w", err)
	}

	return nil
}

// executeReactionResponse adds a reaction to the message
func executeReactionResponse(ctx context.Context, session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	if cfg.Reaction == "" {
		return fmt.Errorf("reaction response requires non-empty reaction")
	}

	if message.ID == "" {
		return fmt.Errorf("message ID is empty, cannot add reaction")
	}

	// Check context before API call
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	err := session.MessageReactionAdd(message.ChannelID, message.ID, cfg.Reaction)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	return nil
}

// BuildEmbed builds a Discord embed from configuration
func BuildEmbed(cfg *config.EmbedConfig) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       cfg.Title,
		Description: cfg.Description,
		Color:       cfg.Color,
	}

	// Add fields
	if len(cfg.Fields) > 0 {
		embed.Fields = make([]*discordgo.MessageEmbedField, len(cfg.Fields))
		for i, field := range cfg.Fields {
			embed.Fields[i] = &discordgo.MessageEmbedField{
				Name:   field.Name,
				Value:  field.Value,
				Inline: field.Inline,
			}
		}
	}

	// Add footer
	if cfg.Footer != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: cfg.Footer,
		}
	}

	// Add timestamp
	if cfg.Timestamp {
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	return embed
}
