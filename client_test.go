package cutline

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestServer builds an httptest.Server with handler, plus a Client that
// points at it with retries disabled (so tests assert the raw behavior).
// Call t.Cleanup to stop the server.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithAPIKey("sk_test_123"),
		WithRetry(0, 10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return srv, c
}

func TestClient_SendsAuthorizationHeader(t *testing.T) {
	t.Parallel()
	var gotAuth string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sport":"soccer-women","division":"d1","count":0,"standings":[]}`))
	})
	if _, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"}); err != nil {
		t.Fatalf("Standings.Get: %v", err)
	}
	if gotAuth != "Bearer sk_test_123" {
		t.Errorf("Authorization = %q, want Bearer sk_test_123", gotAuth)
	}
}

func TestClient_SendsUserAgent(t *testing.T) {
	t.Parallel()
	var gotUA string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sport":"soccer-women","division":"d1","count":0,"standings":[]}`))
	})
	if _, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"}); err != nil {
		t.Fatalf("Standings.Get: %v", err)
	}
	if !strings.HasPrefix(gotUA, "cutline-go-sdk/") {
		t.Errorf("User-Agent = %q, want cutline-go-sdk/* prefix", gotUA)
	}
}

func TestClient_OmitsAuthWhenNoKey(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"sport":"soccer-women","division":"d1","count":0,"standings":[]}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"}); err != nil {
		t.Fatalf("Standings.Get: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization = %q, want empty with no API key", gotAuth)
	}
}

func TestClient_ParsesCutlineErrorEnvelope(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req-abc")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"division is required","code":"BAD_REQUEST"}`))
	})
	_, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err is not *APIError: %v", err)
	}
	if apiErr.Code != "BAD_REQUEST" {
		t.Errorf("Code = %q, want BAD_REQUEST", apiErr.Code)
	}
	if apiErr.Detail != "division is required" {
		t.Errorf("Detail = %q", apiErr.Detail)
	}
	if apiErr.RequestID != "req-abc" {
		t.Errorf("RequestID = %q, want req-abc", apiErr.RequestID)
	}
}

func TestClient_Retries5xxThenSucceeds(t *testing.T) {
	t.Parallel()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sport":"soccer-women","division":"d1","count":0,"standings":[]}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithRetry(2, 1*time.Millisecond))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"}); err != nil {
		t.Fatalf("Standings.Get: %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != 2 {
		t.Errorf("server hits = %d, want 2", got)
	}
}

func TestClient_DoesNotRetry429(t *testing.T) {
	t.Parallel()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited","code":"RATE_LIMIT"}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithRetry(3, 1*time.Millisecond))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"})
	if !IsRateLimited(err) {
		t.Errorf("err = %v, want IsRateLimited", err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Errorf("server hits = %d, want 1 (no retry on 429)", got)
	}
}

func TestClient_DoesNotRetry403(t *testing.T) {
	t.Parallel()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"out of scope","code":"FORBIDDEN"}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithRetry(3, 1*time.Millisecond))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"})
	if !IsOutOfScope(err) {
		t.Errorf("err = %v, want IsOutOfScope", err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Errorf("server hits = %d, want 1 (no retry on 403)", got)
	}
}

func TestClient_RequestHookFiresPerAttempt(t *testing.T) {
	t.Parallel()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sport":"soccer-women","division":"d1","count":0,"standings":[]}`))
	}))
	t.Cleanup(srv.Close)
	var hookCalls int32
	hook := func(_ context.Context, _, _ string) {
		atomic.AddInt32(&hookCalls, 1)
	}
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithRetry(3, 1*time.Millisecond),
		WithRequestHook(hook),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"}); err != nil {
		t.Fatalf("Standings.Get: %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != 3 {
		t.Errorf("server hits = %d, want 3", got)
	}
	if got := atomic.LoadInt32(&hookCalls); got != 3 {
		t.Errorf("hook calls = %d, want 3 (one per attempt including retries)", got)
	}
}

func TestClient_ContextCancelledMidRequest(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithRetry(0, time.Millisecond))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = c.Standings.Get(ctx, StandingsParams{Sport: "soccer-women", Division: "d1"})
	if err == nil {
		t.Fatal("expected ctx deadline error, got nil")
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Errorf("returned in %v; want prompt return on ctx deadline", time.Since(start))
	}
}

func TestClient_InvalidJSONResponseWrapsInvalidResponse(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json at all`))
	})
	_, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"})
	if !errors.Is(err, ErrInvalidResponse) {
		t.Errorf("err = %v, want errors.Is ErrInvalidResponse", err)
	}
}
