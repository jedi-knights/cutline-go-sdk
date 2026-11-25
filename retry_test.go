package cutline

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRetryableStatus(t *testing.T) {
	t.Parallel()
	cases := map[int]bool{
		200:                            false,
		http.StatusBadRequest:          false,
		http.StatusPaymentRequired:     false,
		http.StatusForbidden:           false,
		http.StatusNotFound:            false,
		http.StatusTooManyRequests:     false, // explicit policy: 429 is not retried on the loop
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}
	for status, want := range cases {
		if got := retryableStatus(status); got != want {
			t.Errorf("retryableStatus(%d) = %v, want %v", status, got, want)
		}
	}
}

func TestBackoff_ClampedToMaxBackoff(t *testing.T) {
	t.Parallel()
	d := backoff(MaxRetries, maxRetryBase, 0)
	if d > maxBackoff {
		t.Errorf("backoff(MaxRetries, maxRetryBase, 0) = %v, want ≤ %v", d, maxBackoff)
	}
}

func TestBackoff_JitterInBounds(t *testing.T) {
	t.Parallel()
	base := 200 * time.Millisecond
	// Jitter formula: d = base << attempt, then d/2 + rand[0..d/2]. Range: [d/2, d].
	// For attempt=0, d = base, so result in [base/2, base].
	for i := 0; i < 50; i++ {
		got := backoff(0, base, 0)
		if got < base/2 || got > base {
			t.Errorf("backoff(0, %v, 0) = %v, want in [%v, %v]", base, got, base/2, base)
		}
	}
}

func TestBackoff_RetryAfterRaisesDelay(t *testing.T) {
	t.Parallel()
	base := 100 * time.Millisecond
	retryAfter := 5 * time.Second
	got := backoff(0, base, retryAfter)
	if got < retryAfter {
		t.Errorf("backoff(_, _, %v) = %v, want ≥ %v", retryAfter, got, retryAfter)
	}
	if got > retryAfterCap {
		t.Errorf("backoff with retryAfter = %v, want ≤ cap %v", got, retryAfterCap)
	}
}

func TestBackoff_RetryAfterClampedToCap(t *testing.T) {
	t.Parallel()
	got := backoff(0, 100*time.Millisecond, time.Hour)
	if got != retryAfterCap {
		t.Errorf("backoff(_, _, 1h) = %v, want %v (cap)", got, retryAfterCap)
	}
}

func TestCapRetryAfter(t *testing.T) {
	t.Parallel()
	if got := capRetryAfter(time.Second); got != time.Second {
		t.Errorf("capRetryAfter(1s) = %v, want 1s", got)
	}
	if got := capRetryAfter(time.Hour); got != retryAfterCap {
		t.Errorf("capRetryAfter(1h) = %v, want %v", got, retryAfterCap)
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		header string
		want   time.Duration
	}{
		{"", 0},
		{"0", 0},
		{"-5", 0},
		{"notanumber", 0},
		{"5", 5 * time.Second},
		{"120", 120 * time.Second},
		// HTTP-date form returns 0 (we only support seconds).
		{"Wed, 21 Oct 2026 07:28:00 GMT", 0},
	} {
		h := http.Header{}
		if tc.header != "" {
			h.Set("Retry-After", tc.header)
		}
		got := parseRetryAfter(h)
		if got != tc.want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tc.header, got, tc.want)
		}
	}
}

func TestSleep_ReturnsOnContextCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := sleep(ctx, time.Hour)
	if err == nil {
		t.Fatal("sleep(canceledCtx, 1h) returned nil, want ctx.Err()")
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Errorf("sleep took %v; expected prompt return on cancel", time.Since(start))
	}
}

func TestSleep_TimerFiresBeforeCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	start := time.Now()
	if err := sleep(ctx, 10*time.Millisecond); err != nil {
		t.Fatalf("sleep: %v", err)
	}
	if time.Since(start) < 10*time.Millisecond {
		t.Errorf("sleep returned too fast: %v", time.Since(start))
	}
}
