package cutline

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestWithAPIKey_Empty(t *testing.T) {
	t.Parallel()
	for _, k := range []string{"", "   ", "\t"} {
		if _, err := NewClient(WithAPIKey(k)); err == nil {
			t.Errorf("WithAPIKey(%q): want error, got nil", k)
		}
	}
}

func TestWithAPIKey_Trims(t *testing.T) {
	t.Parallel()
	c, err := NewClient(WithAPIKey("  sk_live_xyz  "))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.apiKey != "sk_live_xyz" {
		t.Errorf("apiKey = %q, want trimmed %q", c.apiKey, "sk_live_xyz")
	}
}

func TestWithBaseURL_RejectsMalformed(t *testing.T) {
	t.Parallel()
	for _, u := range []string{"", "   ", "ftp://foo", "://nohost", "notaurl"} {
		if _, err := NewClient(WithBaseURL(u)); err == nil {
			t.Errorf("WithBaseURL(%q): want error, got nil", u)
		}
	}
}

func TestWithBaseURL_StripsTrailingSlash(t *testing.T) {
	t.Parallel()
	c, err := NewClient(WithBaseURL("https://example.test/"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if strings.HasSuffix(c.baseURL, "/") {
		t.Errorf("baseURL = %q, want no trailing slash", c.baseURL)
	}
}

func TestWithHTTPClient_RejectsNil(t *testing.T) {
	t.Parallel()
	if _, err := NewClient(WithHTTPClient(nil)); err == nil {
		t.Error("WithHTTPClient(nil): want error, got nil")
	}
}

func TestWithHTTPClient_Replaces(t *testing.T) {
	t.Parallel()
	hc := &http.Client{Timeout: 42 * time.Second}
	c, err := NewClient(WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.httpClient != hc {
		t.Error("httpClient was not replaced")
	}
}

func TestWithUserAgent_RejectsEmpty(t *testing.T) {
	t.Parallel()
	for _, ua := range []string{"", "   "} {
		if _, err := NewClient(WithUserAgent(ua)); err == nil {
			t.Errorf("WithUserAgent(%q): want error, got nil", ua)
		}
	}
}

func TestWithRetry_RejectsNegative(t *testing.T) {
	t.Parallel()
	if _, err := NewClient(WithRetry(-1, time.Second)); err == nil {
		t.Error("WithRetry(-1, _): want error")
	}
	if _, err := NewClient(WithRetry(2, -time.Second)); err == nil {
		t.Error("WithRetry(_, -1s): want error")
	}
}

func TestWithRetry_ClampsToMaxRetries(t *testing.T) {
	t.Parallel()
	c, err := NewClient(WithRetry(MaxRetries+99, time.Second))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.maxRetries != MaxRetries {
		t.Errorf("maxRetries = %d, want clamp to %d", c.maxRetries, MaxRetries)
	}
}

func TestWithRetry_ClampsBase(t *testing.T) {
	t.Parallel()
	c, err := NewClient(WithRetry(1, maxRetryBase+time.Hour))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.retryBase != maxRetryBase {
		t.Errorf("retryBase = %v, want clamp to %v", c.retryBase, maxRetryBase)
	}
}

func TestWithRequestHook_RejectsNil(t *testing.T) {
	t.Parallel()
	if _, err := NewClient(WithRequestHook(nil)); err == nil {
		t.Error("WithRequestHook(nil): want error")
	}
}

func TestNewClient_OptionOrderFirstErrorWins(t *testing.T) {
	t.Parallel()
	_, err := NewClient(
		WithAPIKey("ok"),
		WithBaseURL(""), // this one fails
		WithRetry(-1, 0),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "base URL") {
		t.Errorf("expected base URL error first, got %v", err)
	}
}

func TestNewClient_Defaults(t *testing.T) {
	t.Parallel()
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.httpClient == nil {
		t.Error("httpClient is nil")
	}
	if c.maxRetries != defaultMaxRetries {
		t.Errorf("maxRetries = %d, want %d", c.maxRetries, defaultMaxRetries)
	}
	if c.retryBase != defaultRetryBase {
		t.Errorf("retryBase = %v, want %v", c.retryBase, defaultRetryBase)
	}
	if !strings.HasPrefix(c.userAgent, "cutline-go-sdk/") {
		t.Errorf("userAgent = %q, want cutline-go-sdk/* prefix", c.userAgent)
	}
	for name, svc := range map[string]any{
		"Standings":   c.Standings,
		"Scoreboard":  c.Scoreboard,
		"Conferences": c.Conferences,
		"Teams":       c.Teams,
		"Programs":    c.Programs,
		"People":      c.People,
		"Brackets":    c.Brackets,
		"Broadcasts":  c.Broadcasts,
	} {
		if svc == nil {
			t.Errorf("Client.%s is nil", name)
		}
	}
}

func TestNewClient_PropagatesOptionError(t *testing.T) {
	t.Parallel()
	optErr := errors.New("bad option")
	badOption := func(_ *Client) error { return optErr }
	_, err := NewClient(badOption)
	if !errors.Is(err, optErr) {
		t.Errorf("err = %v, want wraps %v", err, optErr)
	}
}
