// Package services holds the backend's stateful collaborators: the database
// handle, JWT signing, the chat WebSocket hub, third-party API clients
// (Spotify, Steam, Gitea, Claude) and the email-sync pipeline. Handlers and
// GraphQL resolvers reach them through handlers.Store.
package services

import (
	"sync"
	"time"
)

// RateLimiter is an in-memory sliding-window limiter keyed by an arbitrary
// string (in practice the client IP). It is process-local: restarting the
// backend clears every window, and a multi-replica deployment would limit
// per replica rather than globally. Nginx applies its own rate limits in
// front of this, so this layer is a second line of defence, not the only one.
type RateLimiter struct {
	mu sync.Mutex
	// attempts holds the timestamps of recent allowed attempts per key.
	// Keys are never evicted, so memory grows with the number of distinct
	// client IPs seen since start-up; acceptable for a personal site, but
	// worth knowing before reusing this elsewhere.
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

// NewRateLimiter returns a limiter allowing at most max calls to Allow per
// key within any window-length period.
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}
}

// Allow records an attempt for key and reports whether it is within the
// limit. A rejected attempt is deliberately NOT recorded, so a caller that
// keeps hammering does not extend its own lockout indefinitely.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter in place: slicing to [:0] reuses the existing backing array, so
	// the append below overwrites the old timestamps instead of allocating a
	// new slice on every request. Safe here because we only ever write an
	// element at an index we have already read past. A missing key yields a
	// nil slice, and nil[:0] is legal and yields nil.
	valid := rl.attempts[key][:0]
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.max {
		rl.attempts[key] = valid
		return false
	}

	rl.attempts[key] = append(valid, now)
	return true
}
