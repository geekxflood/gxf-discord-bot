package action_test

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/gxf-discord-bot/internal/testutil"
	"github.com/geekxflood/gxf-discord-bot/pkg/action"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewManager_Success(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Prefix: "!",
		},
		Actions: []config.ActionConfig{
			{
				Name: "ping",
				Type: "command",
				Trigger: config.TriggerConfig{
					Command: "ping",
				},
				Response: config.ResponseConfig{
					Type:    "text",
					Content: "Pong!",
				},
			},
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()

	mgr, err := action.NewManager(cfg, logger)

	require.NoError(t, err)
	require.NotNil(t, mgr)
}

func TestNewManager_NoActions(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Prefix: "!",
		},
		Actions: []config.ActionConfig{},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()

	mgr, err := action.NewManager(cfg, logger)

	require.NoError(t, err)
	require.NotNil(t, mgr)
}

func TestCommandHandler_Match(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		command     string
		content     string
		shouldMatch bool
	}{
		{
			name:        "exact match",
			prefix:      "!",
			command:     "ping",
			content:     "!ping",
			shouldMatch: true,
		},
		{
			name:        "match with trailing space",
			prefix:      "!",
			command:     "ping",
			content:     "!ping ",
			shouldMatch: true,
		},
		{
			name:        "match with arguments",
			prefix:      "!",
			command:     "echo",
			content:     "!echo hello world",
			shouldMatch: true,
		},
		{
			name:        "no match - wrong prefix",
			prefix:      "!",
			command:     "ping",
			content:     "?ping",
			shouldMatch: false,
		},
		{
			name:        "no match - wrong command",
			prefix:      "!",
			command:     "ping",
			content:     "!pong",
			shouldMatch: false,
		},
		{
			name:        "no match - no prefix",
			prefix:      "!",
			command:     "ping",
			content:     "ping",
			shouldMatch: false,
		},
		{
			name:        "case insensitive match",
			prefix:      "!",
			command:     "ping",
			content:     "!PING",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := action.NewCommandHandler(tt.prefix, tt.command)
			matches := handler.Matches(tt.content)
			assert.Equal(t, tt.shouldMatch, matches)
		})
	}
}

func TestCommandHandler_ExtractArgs(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		command  string
		content  string
		expected []string
	}{
		{
			name:     "no arguments",
			prefix:   "!",
			command:  "ping",
			content:  "!ping",
			expected: []string{},
		},
		{
			name:     "single argument",
			prefix:   "!",
			command:  "echo",
			content:  "!echo hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple arguments",
			prefix:   "!",
			command:  "echo",
			content:  "!echo hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "arguments with extra spaces",
			prefix:   "!",
			command:  "echo",
			content:  "!echo  hello   world  ",
			expected: []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := action.NewCommandHandler(tt.prefix, tt.command)
			args := handler.ExtractArgs(tt.content)
			assert.Equal(t, tt.expected, args)
		})
	}
}

func TestMessageHandler_Match(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		content     string
		shouldMatch bool
	}{
		{
			name:        "simple match",
			pattern:     "hello",
			content:     "hello",
			shouldMatch: true,
		},
		{
			name:        "case insensitive match",
			pattern:     "(?i)hello",
			content:     "HELLO",
			shouldMatch: true,
		},
		{
			name:        "regex pattern match",
			pattern:     "^hello\\s+bot$",
			content:     "hello bot",
			shouldMatch: true,
		},
		{
			name:        "no match",
			pattern:     "hello",
			content:     "goodbye",
			shouldMatch: false,
		},
		{
			name:        "partial match in string",
			pattern:     "hello",
			content:     "well hello there",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := action.NewMessageHandler(tt.pattern)
			require.NoError(t, err)

			matches := handler.Matches(tt.content)
			assert.Equal(t, tt.shouldMatch, matches)
		})
	}
}

func TestMessageHandler_InvalidPattern(t *testing.T) {
	handler, err := action.NewMessageHandler("[invalid(")
	assert.Error(t, err)
	assert.Nil(t, handler)
}

func TestReactionHandler_Match(t *testing.T) {
	tests := []struct {
		name        string
		emoji       string
		reaction    string
		shouldMatch bool
	}{
		{
			name:        "exact match",
			emoji:       "👍",
			reaction:    "👍",
			shouldMatch: true,
		},
		{
			name:        "emoji name match",
			emoji:       "thumbsup",
			reaction:    "thumbsup",
			shouldMatch: true,
		},
		{
			name:        "no match",
			emoji:       "👍",
			reaction:    "👎",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := action.NewReactionHandler(tt.emoji)
			matches := handler.Matches(tt.reaction)
			assert.Equal(t, tt.shouldMatch, matches)
		})
	}
}

func TestManager_HandleMessage(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Prefix: "!",
		},
		Actions: []config.ActionConfig{
			{
				Name: "ping",
				Type: "command",
				Trigger: config.TriggerConfig{
					Command: "ping",
				},
				Response: config.ResponseConfig{
					Type:    "text",
					Content: "Pong!",
				},
			},
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	mgr, err := action.NewManager(cfg, logger)
	require.NoError(t, err)

	session := &testutil.MockDiscordSession{}
	session.On("ChannelMessageSend", "channel123", "Pong!").Return(&discordgo.Message{}, nil)

	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content:   "!ping",
			ChannelID: "channel123",
			Author: &discordgo.User{
				ID:       "123",
				Username: "testuser",
				Bot:      false,
			},
		},
	}

	ctx := context.Background()
	err = mgr.HandleMessage(ctx, session, message)

	assert.NoError(t, err)
	session.AssertExpectations(t)
}

func TestManager_HandleMessage_NoMatch(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Prefix: "!",
		},
		Actions: []config.ActionConfig{
			{
				Name: "ping",
				Type: "command",
				Trigger: config.TriggerConfig{
					Command: "ping",
				},
				Response: config.ResponseConfig{
					Type:    "text",
					Content: "Pong!",
				},
			},
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	mgr, err := action.NewManager(cfg, logger)
	require.NoError(t, err)

	session := &testutil.MockDiscordSession{}
	// No expectations - message won't match

	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content:   "hello",
			ChannelID: "channel123",
			Author: &discordgo.User{
				ID:       "123",
				Username: "testuser",
				Bot:      false,
			},
		},
	}

	ctx := context.Background()
	err = mgr.HandleMessage(ctx, session, message)

	assert.NoError(t, err)
}

func TestManager_GetActions(t *testing.T) {
	cfg := &config.Config{
		Bot: config.BotConfig{
			Prefix: "!",
		},
		Actions: []config.ActionConfig{
			{Name: "ping", Type: "command"},
			{Name: "hello", Type: "message"},
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()

	mgr, err := action.NewManager(cfg, logger)
	require.NoError(t, err)

	actions := mgr.GetActions()
	assert.Len(t, actions, 2)
}
