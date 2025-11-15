package response_test

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/gxf-discord-bot/internal/testutil"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/geekxflood/gxf-discord-bot/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExecuteTextResponse(t *testing.T) {
	cfg := config.ResponseConfig{
		Type:    "text",
		Content: "Hello, World!",
	}

	logger := &testutil.MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session := &testutil.MockDiscordSession{}
	session.On("ChannelMessageSend", "channel123", "Hello, World!").Return(&discordgo.Message{}, nil)

	message := &discordgo.Message{
		ChannelID: "channel123",
		Author: &discordgo.User{
			ID:       "user123",
			Username: "testuser",
		},
	}

	ctx := context.Background()
	err := response.Execute(ctx, session, message, cfg, logger)

	require.NoError(t, err)
	session.AssertExpectations(t)
}

func TestExecuteEmbedResponse(t *testing.T) {
	cfg := config.ResponseConfig{
		Type: "embed",
		Embed: &config.EmbedConfig{
			Title:       "Test Title",
			Description: "Test Description",
			Color:       0x00FF00,
			Fields: []config.EmbedField{
				{
					Name:   "Field 1",
					Value:  "Value 1",
					Inline: true,
				},
			},
			Footer:    "Test Footer",
			Timestamp: true,
		},
	}

	logger := &testutil.MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session := &testutil.MockDiscordSession{}
	session.On("ChannelMessageSendEmbed", "channel123", mock.MatchedBy(func(embed *discordgo.MessageEmbed) bool {
		return embed.Title == "Test Title" &&
			embed.Description == "Test Description" &&
			embed.Color == 0x00FF00 &&
			len(embed.Fields) == 1 &&
			embed.Footer.Text == "Test Footer"
	})).Return(&discordgo.Message{}, nil)

	message := &discordgo.Message{
		ChannelID: "channel123",
		Author: &discordgo.User{
			ID:       "user123",
			Username: "testuser",
		},
	}

	ctx := context.Background()
	err := response.Execute(ctx, session, message, cfg, logger)

	require.NoError(t, err)
	session.AssertExpectations(t)
}

func TestExecuteDMResponse(t *testing.T) {
	cfg := config.ResponseConfig{
		Type:    "dm",
		Content: "Private message",
	}

	logger := &testutil.MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session := &testutil.MockDiscordSession{}
	// First, UserChannelCreate to get DM channel
	session.On("UserChannelCreate", "user123").Return(&discordgo.Channel{ID: "dm-channel"}, nil)
	// Then send message to DM channel
	session.On("ChannelMessageSend", "dm-channel", "Private message").Return(&discordgo.Message{}, nil)

	message := &discordgo.Message{
		ChannelID: "channel123",
		Author: &discordgo.User{
			ID:       "user123",
			Username: "testuser",
		},
	}

	ctx := context.Background()
	err := response.Execute(ctx, session, message, cfg, logger)

	require.NoError(t, err)
	session.AssertExpectations(t)
}

func TestExecuteReactionResponse(t *testing.T) {
	cfg := config.ResponseConfig{
		Type:     "reaction",
		Reaction: "👍",
	}

	logger := &testutil.MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session := &testutil.MockDiscordSession{}
	session.On("MessageReactionAdd", "channel123", "msg123", "👍").Return(nil)

	message := &discordgo.Message{
		ID:        "msg123",
		ChannelID: "channel123",
		Author: &discordgo.User{
			ID:       "user123",
			Username: "testuser",
		},
	}

	ctx := context.Background()
	err := response.Execute(ctx, session, message, cfg, logger)

	require.NoError(t, err)
	session.AssertExpectations(t)
}

func TestExecuteInvalidResponseType(t *testing.T) {
	cfg := config.ResponseConfig{
		Type:    "invalid",
		Content: "test",
	}

	logger := &testutil.MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session := &testutil.MockDiscordSession{}
	message := &discordgo.Message{
		ChannelID: "channel123",
		Author: &discordgo.User{
			ID:       "user123",
			Username: "testuser",
		},
	}

	ctx := context.Background()
	err := response.Execute(ctx, session, message, cfg, logger)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported response type")
}

func TestBuildEmbed(t *testing.T) {
	embedCfg := &config.EmbedConfig{
		Title:       "Test",
		Description: "Description",
		Color:       0xFF0000,
		Fields: []config.EmbedField{
			{Name: "Field1", Value: "Value1", Inline: true},
			{Name: "Field2", Value: "Value2", Inline: false},
		},
		Footer:    "Footer Text",
		Timestamp: true,
	}

	embed := response.BuildEmbed(embedCfg)

	assert.Equal(t, "Test", embed.Title)
	assert.Equal(t, "Description", embed.Description)
	assert.Equal(t, 0xFF0000, embed.Color)
	assert.Len(t, embed.Fields, 2)
	assert.Equal(t, "Field1", embed.Fields[0].Name)
	assert.Equal(t, "Value1", embed.Fields[0].Value)
	assert.True(t, embed.Fields[0].Inline)
	assert.Equal(t, "Footer Text", embed.Footer.Text)
	assert.NotEmpty(t, embed.Timestamp)
}

func TestBuildEmbed_NoTimestamp(t *testing.T) {
	embedCfg := &config.EmbedConfig{
		Title:     "Test",
		Timestamp: false,
	}

	embed := response.BuildEmbed(embedCfg)

	assert.Equal(t, "Test", embed.Title)
	assert.Empty(t, embed.Timestamp)
}

func TestExecuteTextResponse_EmptyContent(t *testing.T) {
	cfg := config.ResponseConfig{
		Type:    "text",
		Content: "",
	}

	logger := &testutil.MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session := &testutil.MockDiscordSession{}
	message := &discordgo.Message{
		ChannelID: "channel123",
		Author: &discordgo.User{
			ID:       "user123",
			Username: "testuser",
		},
	}

	ctx := context.Background()
	err := response.Execute(ctx, session, message, cfg, logger)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty content")
}

func TestExecuteEmbedResponse_NilEmbed(t *testing.T) {
	cfg := config.ResponseConfig{
		Type:  "embed",
		Embed: nil,
	}

	logger := &testutil.MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session := &testutil.MockDiscordSession{}
	message := &discordgo.Message{
		ChannelID: "channel123",
		Author: &discordgo.User{
			ID:       "user123",
			Username: "testuser",
		},
	}

	ctx := context.Background()
	err := response.Execute(ctx, session, message, cfg, logger)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "embed config is nil")
}
