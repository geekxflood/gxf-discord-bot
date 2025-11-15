// Package ratelimit provides rate limiting functionality for Discord bot actions.
package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/geekxflood/common/logging"
)

// Limiter manages rate limits for users, channels, guilds, and globally
type Limiter struct {
	logger logging.Logger

	// User rate limits
	userLimit    int
	userWindow   time.Duration
	userBuckets  map[string]*bucket
	userMu       sync.RWMutex

	// Channel rate limits
	channelLimit   int
	channelWindow  time.Duration
	channelBuckets map[string]*bucket
	channelMu      sync.RWMutex

	// Guild rate limits
	guildLimit   int
	guildWindow  time.Duration
	guildBuckets map[string]*bucket
	guildMu      sync.RWMutex

	// Global rate limit
	globalLimit  int
	globalWindow time.Duration
	globalBucket *bucket
	globalMu     sync.RWMutex

	// Cleanup
	cleanupStop chan struct{}
	cleanupMu   sync.Mutex
}

type bucket struct {
	tokens    int
	maxTokens int
	window    time.Duration
	lastReset time.Time
	mu        sync.Mutex
}

// New creates a new rate limiter
func New(logger logging.Logger) *Limiter {
	logger.Info("Creating new rate limiter")

	return &Limiter{
		logger:         logger,
		userBuckets:    make(map[string]*bucket),
		channelBuckets: make(map[string]*bucket),
		guildBuckets:   make(map[string]*bucket),
	}
}

// SetUserLimit configures per-user rate limiting
func (l *Limiter) SetUserLimit(limit int, window time.Duration) {
	l.userMu.Lock()
	defer l.userMu.Unlock()

	l.userLimit = limit
	l.userWindow = window
	l.logger.Debug("User rate limit configured", "limit", limit, "window", window)
}

// SetChannelLimit configures per-channel rate limiting
func (l *Limiter) SetChannelLimit(limit int, window time.Duration) {
	l.channelMu.Lock()
	defer l.channelMu.Unlock()

	l.channelLimit = limit
	l.channelWindow = window
	l.logger.Debug("Channel rate limit configured", "limit", limit, "window", window)
}

// SetGuildLimit configures per-guild rate limiting
func (l *Limiter) SetGuildLimit(limit int, window time.Duration) {
	l.guildMu.Lock()
	defer l.guildMu.Unlock()

	l.guildLimit = limit
	l.guildWindow = window
	l.logger.Debug("Guild rate limit configured", "limit", limit, "window", window)
}

// SetGlobalLimit configures global rate limiting
func (l *Limiter) SetGlobalLimit(limit int, window time.Duration) {
	l.globalMu.Lock()
	defer l.globalMu.Unlock()

	l.globalLimit = limit
	l.globalWindow = window
	l.globalBucket = &bucket{
		tokens:    limit,
		maxTokens: limit,
		window:    window,
		lastReset: time.Now(),
	}
	l.logger.Debug("Global rate limit configured", "limit", limit, "window", window)
}

// AllowUser checks if a user is allowed to make a request
func (l *Limiter) AllowUser(userID string) bool {
	l.userMu.Lock()
	defer l.userMu.Unlock()

	// If no limit configured, allow
	if l.userLimit == 0 {
		return true
	}

	// Get or create bucket
	b, exists := l.userBuckets[userID]
	if !exists {
		b = &bucket{
			tokens:    l.userLimit,
			maxTokens: l.userLimit,
			window:    l.userWindow,
			lastReset: time.Now(),
		}
		l.userBuckets[userID] = b
	}

	return b.allow()
}

// AllowChannel checks if a channel is allowed to make a request
func (l *Limiter) AllowChannel(channelID string) bool {
	l.channelMu.Lock()
	defer l.channelMu.Unlock()

	// If no limit configured, allow
	if l.channelLimit == 0 {
		return true
	}

	// Get or create bucket
	b, exists := l.channelBuckets[channelID]
	if !exists {
		b = &bucket{
			tokens:    l.channelLimit,
			maxTokens: l.channelLimit,
			window:    l.channelWindow,
			lastReset: time.Now(),
		}
		l.channelBuckets[channelID] = b
	}

	return b.allow()
}

// AllowGuild checks if a guild is allowed to make a request
func (l *Limiter) AllowGuild(guildID string) bool {
	l.guildMu.Lock()
	defer l.guildMu.Unlock()

	// If no limit configured, allow
	if l.guildLimit == 0 {
		return true
	}

	// Get or create bucket
	b, exists := l.guildBuckets[guildID]
	if !exists {
		b = &bucket{
			tokens:    l.guildLimit,
			maxTokens: l.guildLimit,
			window:    l.guildWindow,
			lastReset: time.Now(),
		}
		l.guildBuckets[guildID] = b
	}

	return b.allow()
}

// AllowGlobal checks if a global request is allowed
func (l *Limiter) AllowGlobal() bool {
	l.globalMu.Lock()
	defer l.globalMu.Unlock()

	// If no limit configured, allow
	if l.globalLimit == 0 {
		return true
	}

	if l.globalBucket == nil {
		return true
	}

	return l.globalBucket.allow()
}

// Allow checks all applicable rate limits
func (l *Limiter) Allow(userID, channelID, guildID string) bool {
	// Check all limits - all must pass
	if !l.AllowUser(userID) {
		l.logger.Warn("User rate limit exceeded", "userID", userID)
		return false
	}

	if !l.AllowChannel(channelID) {
		l.logger.Warn("Channel rate limit exceeded", "channelID", channelID)
		return false
	}

	if !l.AllowGuild(guildID) {
		l.logger.Warn("Guild rate limit exceeded", "guildID", guildID)
		return false
	}

	if !l.AllowGlobal() {
		l.logger.Warn("Global rate limit exceeded")
		return false
	}

	return true
}

// ResetUser resets the rate limit for a specific user
func (l *Limiter) ResetUser(userID string) {
	l.userMu.Lock()
	defer l.userMu.Unlock()

	delete(l.userBuckets, userID)
	l.logger.Debug("User rate limit reset", "userID", userID)
}

// GetUserRemaining returns the remaining requests for a user
func (l *Limiter) GetUserRemaining(userID string) int {
	l.userMu.RLock()
	defer l.userMu.RUnlock()

	if l.userLimit == 0 {
		return -1 // unlimited
	}

	b, exists := l.userBuckets[userID]
	if !exists {
		return l.userLimit
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.reset()
	return b.tokens
}

// Cleanup removes expired rate limit buckets
func (l *Limiter) Cleanup() {
	l.logger.Debug("Running rate limit cleanup")

	now := time.Now()

	// Clean user buckets
	l.userMu.Lock()
	for id, b := range l.userBuckets {
		if now.Sub(b.lastReset) > b.window {
			delete(l.userBuckets, id)
		}
	}
	l.userMu.Unlock()

	// Clean channel buckets
	l.channelMu.Lock()
	for id, b := range l.channelBuckets {
		if now.Sub(b.lastReset) > b.window {
			delete(l.channelBuckets, id)
		}
	}
	l.channelMu.Unlock()

	// Clean guild buckets
	l.guildMu.Lock()
	for id, b := range l.guildBuckets {
		if now.Sub(b.lastReset) > b.window {
			delete(l.guildBuckets, id)
		}
	}
	l.guildMu.Unlock()
}

// StartCleanup starts automatic cleanup of expired buckets
func (l *Limiter) StartCleanup(interval time.Duration) error {
	l.cleanupMu.Lock()
	defer l.cleanupMu.Unlock()

	if l.cleanupStop != nil {
		return fmt.Errorf("cleanup already running")
	}

	l.cleanupStop = make(chan struct{})
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				l.Cleanup()
			case <-l.cleanupStop:
				ticker.Stop()
				return
			}
		}
	}()

	l.logger.Info("Rate limit cleanup started", "interval", interval)
	return nil
}

// StopCleanup stops automatic cleanup
func (l *Limiter) StopCleanup() {
	l.cleanupMu.Lock()
	defer l.cleanupMu.Unlock()

	if l.cleanupStop != nil {
		close(l.cleanupStop)
		l.cleanupStop = nil
		l.logger.Info("Rate limit cleanup stopped")
	}
}

// allow checks if the bucket allows a request
func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Reset if window has passed
	b.reset()

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// reset resets the bucket if the window has passed
func (b *bucket) reset() {
	if time.Since(b.lastReset) >= b.window {
		b.tokens = b.maxTokens
		b.lastReset = time.Now()
	}
}
