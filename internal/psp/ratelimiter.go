package psp

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string][]time.Time
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

func (r *RateLimiter) Allow(vpa string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	times := r.buckets[vpa]
	valid := make([]time.Time, 0)
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= r.limit {
		fmt.Println("psp: rate limit exceeded for", vpa, "requests in window:", len(valid))
		return errors.New("rate limit exceeded for " + vpa)
	}

	r.buckets[vpa] = append(valid, now)
	return nil
}
