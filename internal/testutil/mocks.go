// Package testutil provides testing utilities and mocks.
package testutil

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/common/logging"
	"github.com/stretchr/testify/mock"
)

// MockDiscordSession is a mock implementation of Discord session for testing
type MockDiscordSession struct {
	mock.Mock
	OpenCalled  bool
	CloseCalled bool
	Handlers    []interface{}
}

// Open mocks the Open method
func (m *MockDiscordSession) Open() error {
	m.OpenCalled = true
	args := m.Called()
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockDiscordSession) Close() error {
	m.CloseCalled = true
	args := m.Called()
	return args.Error(0)
}

// AddHandler mocks adding event handlers
func (m *MockDiscordSession) AddHandler(handler interface{}) func() {
	m.Handlers = append(m.Handlers, handler)
	return func() {}
}

// UpdateStatusComplex mocks updating bot status
func (m *MockDiscordSession) UpdateStatusComplex(data discordgo.UpdateStatusData) error {
	args := m.Called(data)
	return args.Error(0)
}

// ChannelMessageSend mocks sending a message to a channel
func (m *MockDiscordSession) ChannelMessageSend(channelID, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	args := m.Called(channelID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discordgo.Message), args.Error(1)
}

// ChannelMessageSendEmbed mocks sending an embed to a channel
func (m *MockDiscordSession) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	args := m.Called(channelID, embed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discordgo.Message), args.Error(1)
}

// UserChannelCreate mocks creating a DM channel with a user
func (m *MockDiscordSession) UserChannelCreate(userID string, options ...discordgo.RequestOption) (*discordgo.Channel, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discordgo.Channel), args.Error(1)
}

// MessageReactionAdd mocks adding a reaction to a message
func (m *MockDiscordSession) MessageReactionAdd(channelID, messageID, emojiID string, options ...discordgo.RequestOption) error {
	args := m.Called(channelID, messageID, emojiID)
	return args.Error(0)
}

// ChannelMessage mocks retrieving a message from a channel
func (m *MockDiscordSession) ChannelMessage(channelID, messageID string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	args := m.Called(channelID, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discordgo.Message), args.Error(1)
}

// MockLogger is a mock implementation of logging.Logger
type MockLogger struct {
	mock.Mock
	InfoMessages  []string
	ErrorMessages []string
	DebugMessages []string
}

// Info logs an info message
func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.InfoMessages = append(m.InfoMessages, msg)
	m.Called(msg, keysAndValues)
}

// Error logs an error message
func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, msg)
	m.Called(msg, keysAndValues)
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.DebugMessages = append(m.DebugMessages, msg)
	m.Called(msg, keysAndValues)
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

// With creates a logger with additional context
func (m *MockLogger) With(keysAndValues ...interface{}) logging.Logger {
	m.Called(keysAndValues)
	return m
}

// InfoContext logs an info message with context
func (m *MockLogger) InfoContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.InfoMessages = append(m.InfoMessages, msg)
	m.Called(ctx, msg, keysAndValues)
}

// ErrorContext logs an error message with context
func (m *MockLogger) ErrorContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, msg)
	m.Called(ctx, msg, keysAndValues)
}

// DebugContext logs a debug message with context
func (m *MockLogger) DebugContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.DebugMessages = append(m.DebugMessages, msg)
	m.Called(ctx, msg, keysAndValues)
}

// WarnContext logs a warning message with context
func (m *MockLogger) WarnContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.Called(ctx, msg, keysAndValues)
}
