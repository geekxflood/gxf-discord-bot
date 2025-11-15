package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

// checkConditions checks if all conditions are met for an action
func (m *ActionManager) checkConditions(ctx context.Context, s *discordgo.Session, msg *discordgo.MessageCreate, action *Action) bool {
	// Check channel restrictions
	if len(action.Trigger.Channels) > 0 {
		found := false
		for _, ch := range action.Trigger.Channels {
			if ch == msg.ChannelID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check guild restrictions
	if len(action.Trigger.Guilds) > 0 {
		found := false
		for _, g := range action.Trigger.Guilds {
			if g == msg.GuildID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check auth requirement
	if action.RequireAuth {
		if !m.authMgr.Enabled() {
			m.logger.Warn("Action requires auth but auth is not enabled", "action", action.Name)
			return false
		}

		if !m.authMgr.IsAuthenticated(msg.Author.ID) {
			// Send auth URL
			authURL := m.authMgr.GetAuthURL(msg.Author.ID)
			_, _ = s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("This command requires authentication. Please authenticate here: %s", authURL))
			return false
		}
	}

	// Check other conditions
	for _, condition := range action.Conditions {
		if !m.checkCondition(ctx, s, msg, condition) {
			return false
		}
	}

	return true
}

// checkCondition checks a single condition
func (m *ActionManager) checkCondition(_ context.Context, s *discordgo.Session, msg *discordgo.MessageCreate, condition *Condition) bool {
	switch condition.Type {
	case "channel":
		return m.compareValue(msg.ChannelID, condition.Value, condition.Operator)

	case "user":
		return m.compareValue(msg.Author.ID, condition.Value, condition.Operator)

	case "role":
		if msg.GuildID == "" {
			return false
		}
		member, err := s.GuildMember(msg.GuildID, msg.Author.ID)
		if err != nil {
			m.logger.Error("Failed to get guild member", "error", err)
			return false
		}

		hasRole := false
		for _, roleID := range member.Roles {
			if roleID == condition.Value {
				hasRole = true
				break
			}
		}
		return hasRole

	case "permission":
		if msg.GuildID == "" {
			return false
		}
		perms, err := s.UserChannelPermissions(msg.Author.ID, msg.ChannelID)
		if err != nil {
			m.logger.Error("Failed to get user permissions", "error", err)
			return false
		}

		// Check if user has the permission (simplified)
		return perms > 0

	default:
		m.logger.Warn("Unknown condition type", "type", condition.Type)
		return false
	}
}

// compareValue compares values based on operator
func (m *ActionManager) compareValue(actual, expected, operator string) bool {
	switch operator {
	case "equals":
		return actual == expected
	case "not":
		return actual != expected
	default:
		return actual == expected
	}
}

// checkRateLimit checks if rate limit is exceeded
func (m *ActionManager) checkRateLimit(userID, channelID, guildID string, action *Action) bool {
	if action.RateLimit == nil {
		return true
	}

	var key string
	switch action.RateLimit.Scope {
	case "user":
		key = fmt.Sprintf("%s:%s", action.Name, userID)
	case "channel":
		key = fmt.Sprintf("%s:%s", action.Name, channelID)
	case "guild":
		key = fmt.Sprintf("%s:%s", action.Name, guildID)
	case "global":
		key = action.Name
	default:
		key = fmt.Sprintf("%s:%s", action.Name, userID)
	}

	return m.rateLimiter.Allow(key, action.RateLimit.Requests, time.Duration(action.RateLimit.Window)*time.Second)
}

// executeAction executes an action
func (m *ActionManager) executeAction(ctx context.Context, s *discordgo.Session, msg *discordgo.MessageCreate, action *Action) error {
	m.logger.Debug("Executing action", "action", action.Name, "user", msg.Author.ID)

	if err := m.sendResponse(ctx, s, msg.ChannelID, msg.Author.ID, action.Response); err != nil {
		return err
	}

	// Delete trigger message if configured
	if action.Response.DeleteAfter > 0 {
		time.AfterFunc(time.Duration(action.Response.DeleteAfter)*time.Second, func() {
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		})
	}

	return nil
}

// sendResponse sends a response based on configuration
func (m *ActionManager) sendResponse(ctx context.Context, s *discordgo.Session, channelID, userID string, response *Response) error {
	switch response.Type {
	case "text":
		_, err := s.ChannelMessageSend(channelID, response.Content)
		return err

	case "embed":
		if response.Embed != nil {
			embed := m.buildEmbed(response.Embed)
			_, err := s.ChannelMessageSendEmbed(channelID, embed)
			return err
		}
		return fmt.Errorf("embed configuration is missing")

	case "reaction":
		// This requires message ID which we don't have in this context
		return fmt.Errorf("reaction type not supported in this context")

	case "dm":
		if userID == "" {
			return fmt.Errorf("user ID required for DM")
		}
		channel, err := s.UserChannelCreate(userID)
		if err != nil {
			return fmt.Errorf("failed to create DM channel: %w", err)
		}

		if response.Embed != nil {
			embed := m.buildEmbed(response.Embed)
			_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
		} else {
			_, err = s.ChannelMessageSend(channel.ID, response.Content)
		}
		return err

	case "http":
		if response.HTTP != nil {
			return m.sendHTTPRequest(ctx, response.HTTP)
		}
		return fmt.Errorf("HTTP configuration is missing")

	case "webhook":
		if response.WebhookURL != "" {
			return m.sendWebhook(ctx, response.WebhookURL, response.Content)
		}
		return fmt.Errorf("webhook URL is missing")

	default:
		return fmt.Errorf("unknown response type: %s", response.Type)
	}
}

// buildEmbed builds a Discord embed from configuration
func (m *ActionManager) buildEmbed(cfg *Embed) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       cfg.Title,
		Description: cfg.Description,
		Color:       cfg.Color,
	}

	if cfg.Footer != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: cfg.Footer,
		}
	}

	if cfg.Image != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: cfg.Image,
		}
	}

	if cfg.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: cfg.Thumbnail,
		}
	}

	if cfg.Author != nil {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    cfg.Author.Name,
			IconURL: cfg.Author.IconURL,
			URL:     cfg.Author.URL,
		}
	}

	if cfg.Timestamp {
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	for _, field := range cfg.Fields {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   field.Name,
			Value:  field.Value,
			Inline: field.Inline,
		})
	}

	return embed
}

// sendHTTPRequest sends an HTTP request
func (m *ActionManager) sendHTTPRequest(ctx context.Context, cfg *HTTPConfig) error {
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, bytes.NewBufferString(cfg.Body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	m.logger.Debug("HTTP request sent", "url", cfg.URL, "status", resp.StatusCode)
	return nil
}

// sendWebhook sends a webhook message
func (m *ActionManager) sendWebhook(ctx context.Context, webhookURL, content string) error {
	body := fmt.Sprintf(`{"content": "%s"}`, content)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}

	return nil
}
