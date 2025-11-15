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

// DiscordSession defines the interface for Discord session methods we need
type DiscordSession interface {
	ChannelMessageSend(channelID, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)
	UserChannelCreate(userID string, options ...discordgo.RequestOption) (*discordgo.Channel, error)
	MessageReactionAdd(channelID, messageID, emojiID string, options ...discordgo.RequestOption) error
}

// Execute executes a response based on the configuration
func Execute(ctx context.Context, session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig, logger logging.Logger) error {
	logger.Debug("Executing response", "type", cfg.Type)

	switch cfg.Type {
	case "text":
		return executeTextResponse(session, message, cfg)
	case "embed":
		return executeEmbedResponse(session, message, cfg)
	case "dm":
		return executeDMResponse(session, message, cfg)
	case "reaction":
		return executeReactionResponse(session, message, cfg)
	default:
		return fmt.Errorf("unsupported response type: %s", cfg.Type)
	}
}

// executeTextResponse sends a text message to the channel
func executeTextResponse(session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	if cfg.Content == "" {
		return fmt.Errorf("text response requires non-empty content")
	}

	_, err := session.ChannelMessageSend(message.ChannelID, cfg.Content)
	if err != nil {
		return fmt.Errorf("failed to send text message: %w", err)
	}

	return nil
}

// executeEmbedResponse sends an embed message to the channel
func executeEmbedResponse(session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	if cfg.Embed == nil {
		return fmt.Errorf("embed response requires non-nil embed config is nil")
	}

	embed := BuildEmbed(cfg.Embed)

	_, err := session.ChannelMessageSendEmbed(message.ChannelID, embed)
	if err != nil {
		return fmt.Errorf("failed to send embed: %w", err)
	}

	return nil
}

// executeDMResponse sends a direct message to the user
func executeDMResponse(session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	// Create DM channel
	channel, err := session.UserChannelCreate(message.Author.ID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel: %w", err)
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
func executeReactionResponse(session DiscordSession, message *discordgo.Message, cfg config.ResponseConfig) error {
	if cfg.Reaction == "" {
		return fmt.Errorf("reaction response requires non-empty reaction")
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
