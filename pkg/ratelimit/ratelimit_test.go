package ratelimit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/geekxflood/gxf-discord-bot/internal/testutil"
	"github.com/geekxflood/gxf-discord-bot/pkg/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewLimiter(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	require.NotNil(t, limiter)
}

func TestLimiter_AllowUser(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	// Configure: 2 requests per 1 second
	limiter.SetUserLimit(2, time.Second)

	userID := "user123"

	// First request should be allowed
	allowed := limiter.AllowUser(userID)
	assert.True(t, allowed)

	// Second request should be allowed
	allowed = limiter.AllowUser(userID)
	assert.True(t, allowed)

	// Third request should be denied (limit reached)
	allowed = limiter.AllowUser(userID)
	assert.False(t, allowed)

	// Wait for window to reset
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	allowed = limiter.AllowUser(userID)
	assert.True(t, allowed)
}

func TestLimiter_AllowChannel(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	// Configure: 3 requests per 1 second
	limiter.SetChannelLimit(3, time.Second)

	channelID := "channel123"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		allowed := limiter.AllowChannel(channelID)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Fourth request should be denied
	allowed := limiter.AllowChannel(channelID)
	assert.False(t, allowed)
}

func TestLimiter_AllowGuild(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	// Configure: 5 requests per 1 second
	limiter.SetGuildLimit(5, time.Second)

	guildID := "guild123"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		allowed := limiter.AllowGuild(guildID)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Sixth request should be denied
	allowed := limiter.AllowGuild(guildID)
	assert.False(t, allowed)
}

func TestLimiter_AllowGlobal(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	// Configure: 10 requests per 1 second globally
	limiter.SetGlobalLimit(10, time.Second)

	// First 10 requests should be allowed
	for i := 0; i < 10; i++ {
		allowed := limiter.AllowGlobal()
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 11th request should be denied
	allowed := limiter.AllowGlobal()
	assert.False(t, allowed)
}

func TestLimiter_CombinedLimits(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	// Configure multiple limits
	limiter.SetUserLimit(5, time.Second)
	limiter.SetChannelLimit(10, time.Second)
	limiter.SetGlobalLimit(20, time.Second)

	userID := "user123"
	channelID := "channel123"
	guildID := "guild123"

	// All checks should pass initially
	allowed := limiter.Allow(userID, channelID, guildID)
	assert.True(t, allowed)
}

func TestLimiter_ResetUser(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)
	limiter.SetUserLimit(1, time.Minute)

	userID := "user123"

	// Use up the limit
	allowed := limiter.AllowUser(userID)
	assert.True(t, allowed)

	allowed = limiter.AllowUser(userID)
	assert.False(t, allowed)

	// Reset the user's limit
	limiter.ResetUser(userID)

	// Should be allowed again
	allowed = limiter.AllowUser(userID)
	assert.True(t, allowed)
}

func TestLimiter_GetUserRemaining(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)
	limiter.SetUserLimit(5, time.Second)

	userID := "user123"

	// Initially should have 5 remaining
	remaining := limiter.GetUserRemaining(userID)
	assert.Equal(t, 5, remaining)

	// Use one
	limiter.AllowUser(userID)

	// Should have 4 remaining
	remaining = limiter.GetUserRemaining(userID)
	assert.Equal(t, 4, remaining)
}

func TestLimiter_Cleanup(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)
	limiter.SetUserLimit(5, 100*time.Millisecond)

	// Use the limiter for multiple users
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user%d", i)
		limiter.AllowUser(userID)
	}

	// Wait for limits to expire
	time.Sleep(150 * time.Millisecond)

	// Run cleanup
	limiter.Cleanup()

	// After cleanup, limits should be reset
	allowed := limiter.AllowUser("user1")
	assert.True(t, allowed)
}

func TestLimiter_StartStopCleanup(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	// Start automatic cleanup
	err := limiter.StartCleanup(100 * time.Millisecond)
	require.NoError(t, err)

	// Stop cleanup
	limiter.StopCleanup()
}

func TestLimiter_NoLimitsConfigured(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)

	// Without configured limits, all requests should be allowed
	allowed := limiter.AllowUser("user123")
	assert.True(t, allowed)

	allowed = limiter.AllowChannel("channel123")
	assert.True(t, allowed)

	allowed = limiter.AllowGuild("guild123")
	assert.True(t, allowed)

	allowed = limiter.AllowGlobal()
	assert.True(t, allowed)
}

func TestLimiter_DifferentUsers(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	limiter := ratelimit.New(logger)
	limiter.SetUserLimit(1, time.Second)

	// User1 uses their limit
	allowed := limiter.AllowUser("user1")
	assert.True(t, allowed)

	allowed = limiter.AllowUser("user1")
	assert.False(t, allowed)

	// User2 should still be allowed (separate limit)
	allowed = limiter.AllowUser("user2")
	assert.True(t, allowed)
}
