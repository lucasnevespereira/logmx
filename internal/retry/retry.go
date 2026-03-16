package retry

import (
	"context"
	"math"
	"time"
)

// Backoff retries fn with exponential backoff.
// Starts at 1s, caps at maxDelay. Resets after a successful call.
func Backoff(ctx context.Context, maxDelay time.Duration, fn func() error) {
	attempt := 0
	for {
		err := fn()
		if err == nil {
			attempt = 0
			continue
		}

		if ctx.Err() != nil {
			return
		}

		delay := time.Duration(math.Min(
			float64(time.Second)*math.Pow(2, float64(attempt)),
			float64(maxDelay),
		))
		attempt++

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}
