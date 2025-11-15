package handlers

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter()

	tests := []struct {
		name        string
		key         string
		maxRequests int
		window      time.Duration
		requests    int
		want        []bool
	}{
		{
			name:        "within limit",
			key:         "test1",
			maxRequests: 5,
			window:      time.Second,
			requests:    3,
			want:        []bool{true, true, true},
		},
		{
			name:        "exceeds limit",
			key:         "test2",
			maxRequests: 2,
			window:      time.Second,
			requests:    4,
			want:        []bool{true, true, false, false},
		},
		{
			name:        "single request limit",
			key:         "test3",
			maxRequests: 1,
			window:      time.Second,
			requests:    3,
			want:        []bool{true, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.requests; i++ {
				got := rl.Allow(tt.key, tt.maxRequests, tt.window)
				if got != tt.want[i] {
					t.Errorf("Allow() request %d = %v, want %v", i, got, tt.want[i])
				}
			}
		})
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter()
	key := "expiry-test"
	maxRequests := 2
	window := 100 * time.Millisecond

	// First two requests should succeed
	if !rl.Allow(key, maxRequests, window) {
		t.Error("First request should be allowed")
	}
	if !rl.Allow(key, maxRequests, window) {
		t.Error("Second request should be allowed")
	}

	// Third request should fail
	if rl.Allow(key, maxRequests, window) {
		t.Error("Third request should be denied")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow(key, maxRequests, window) {
		t.Error("Request after window expiry should be allowed")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter()
	key := "concurrent-test"
	maxRequests := 10
	window := time.Second
	goroutines := 20

	var wg sync.WaitGroup
	var allowedCount int32
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow(key, maxRequests, window) {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowedCount != int32(maxRequests) {
		t.Errorf("Expected %d requests allowed, got %d", maxRequests, allowedCount)
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter()
	maxRequests := 1
	window := time.Second

	// Different keys should have separate limits
	if !rl.Allow("key1", maxRequests, window) {
		t.Error("First request for key1 should be allowed")
	}
	if !rl.Allow("key2", maxRequests, window) {
		t.Error("First request for key2 should be allowed")
	}

	// Second requests should fail
	if rl.Allow("key1", maxRequests, window) {
		t.Error("Second request for key1 should be denied")
	}
	if rl.Allow("key2", maxRequests, window) {
		t.Error("Second request for key2 should be denied")
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter()
	key := "benchmark"
	maxRequests := 100
	window := time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow(key, maxRequests, window)
	}
}

func BenchmarkRateLimiter_AllowConcurrent(b *testing.B) {
	rl := NewRateLimiter()
	maxRequests := 100
	window := time.Second

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "benchmark-concurrent-" + string(rune(i%10))
			rl.Allow(key, maxRequests, window)
			i++
		}
	})
}
