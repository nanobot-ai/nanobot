package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	lock     sync.Mutex
	window   time.Duration
	max      int
}

func NewRateLimiter(window time.Duration, max int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		max:      max,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.lock.Lock()
	defer rl.lock.Unlock()
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	var validRequests []time.Time
	for _, reqTime := range rl.requests[key] {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	if len(validRequests) >= rl.max {
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	return true
}
