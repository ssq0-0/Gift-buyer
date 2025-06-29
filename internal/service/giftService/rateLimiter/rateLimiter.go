package rateLimiter

import (
	"context"
	"sync"
	"time"
)

// rateLimiter implements a token bucket rate limiter for API calls
type rateLimiterImpl struct {
	tokens    chan struct{}
	ticker    *time.Ticker
	maxTokens int
	mu        sync.Mutex
	closed    bool
}

// newRateLimiter creates a new rate limiter with specified rate (requests per second)
func NewRateLimiter(rps int) *rateLimiterImpl {
	var ticker *time.Ticker
	if rps > 0 {
		ticker = time.NewTicker(time.Second / time.Duration(rps))
	} else {
		ticker = time.NewTicker(time.Hour)
	}

	rl := &rateLimiterImpl{
		tokens:    make(chan struct{}, rps),
		ticker:    ticker,
		maxTokens: rps,
	}

	// Заполняем канал начальными токенами
	for i := 0; i < rps; i++ {
		rl.tokens <- struct{}{}
	}

	if rps > 0 {
		go rl.refillTokens()
	}

	return rl
}

func (rl *rateLimiterImpl) Acquire(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *rateLimiterImpl) refillTokens() {
	for range rl.ticker.C {
		rl.mu.Lock()
		if rl.closed {
			rl.mu.Unlock()
			return
		}

		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full, skip
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiterImpl) Close() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.closed {
		rl.closed = true
		rl.ticker.Stop()
		// Не закрываем канал, так как он может использоваться в тестах
	}
}
