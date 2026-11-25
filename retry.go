package cutline

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// Backoff bounds. Delays never exceed maxBackoff, and a server-sent
// Retry-After is honored only up to retryAfterCap, so a misbehaving server
// cannot park a caller for long.
const (
	maxBackoff    = 2 * time.Second
	retryAfterCap = 30 * time.Second
)

// retryableStatus reports whether an HTTP status is worth retrying.
// 402, 403, 404, and 429 are not retried: they are plan / scope / rate-limit
// decisions that will keep failing on an immediate retry.
func retryableStatus(code int) bool {
	switch code {
	case http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

// backoff returns the delay before retry number attempt (0-based): exponential
// from base, capped at maxBackoff, with jitter in [d/2, d] so callers that
// failed together do not retry together. A Retry-After value raises the delay
// (up to retryAfterCap).
func backoff(attempt int, base, retryAfter time.Duration) time.Duration {
	// base <= maxRetryBase and attempt <= MaxRetries, so the shift cannot overflow.
	d := base << min(attempt, MaxRetries)
	d = min(d, maxBackoff)
	if d <= 0 {
		d = base
	}
	d = d/2 + time.Duration(rand.Int64N(int64(d/2)+1)) // jitter; not security-sensitive
	if retryAfter > d {
		d = capRetryAfter(retryAfter)
	}
	return d
}

// capRetryAfter clamps a server-sent Retry-After to retryAfterCap.
func capRetryAfter(d time.Duration) time.Duration {
	if d > retryAfterCap {
		return retryAfterCap
	}
	return d
}

// parseRetryAfter reads a Retry-After header given in seconds. HTTP-date
// values and anything unparsable return 0, which means "use the backoff".
func parseRetryAfter(h http.Header) time.Duration {
	secs, err := strconv.Atoi(h.Get("Retry-After"))
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

// sleep waits for d or until ctx is done, whichever comes first.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
