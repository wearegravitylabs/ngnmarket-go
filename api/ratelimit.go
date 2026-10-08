package api

import (
	"context"
	"sync"
	"time"
)

// limiter is a small token bucket used for the optional client-side rate limit
// (see WithRateLimit). It keeps the SDK stdlib-only.
type limiter struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64
	perSec   float64
	last     time.Time
	now      func() time.Time // injectable for tests
}

func newLimiter(requestsPerMinute, burst int) *limiter {
	if burst < 1 {
		burst = 1
	}
	return &limiter{
		tokens:   float64(burst),
		capacity: float64(burst),
		perSec:   float64(requestsPerMinute) / 60.0,
		last:     time.Now(),
		now:      time.Now,
	}
}

// wait blocks until a token is available or ctx is done.
func (l *limiter) wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := l.now()
		l.tokens += now.Sub(l.last).Seconds() * l.perSec
		if l.tokens > l.capacity {
			l.tokens = l.capacity
		}
		l.last = now
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		need := time.Duration((1 - l.tokens) / l.perSec * float64(time.Second))
		l.mu.Unlock()

		t := time.NewTimer(need)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
}
