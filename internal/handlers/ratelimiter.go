package handlers

import (
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed based on rate limits
func (rl *RateLimiter) Allow(key string, maxRequests int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	// Get existing requests for this key
	requests, exists := rl.requests[key]
	if !exists {
		requests = []time.Time{}
	}

	// Filter out expired requests
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit is exceeded
	if len(validRequests) >= maxRequests {
		rl.requests[key] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// cleanup periodically cleans up old entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-1 * time.Hour)

		for key, requests := range rl.requests {
			validRequests := make([]time.Time, 0)
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}
