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
