package cutline

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Retry limits. They are named constants so the bound on every retry loop is
// discoverable and reviewable.
const (
	// MaxRetries is the largest retry count WithRetry accepts.
	MaxRetries = 10
	// maxRetryBase is the largest base delay WithRetry accepts, which also
	// keeps the exponential shift in backoff from overflowing.
	maxRetryBase = time.Minute
)

// RequestHook observes each request attempt just before it is sent, including
// retries. Use it for metrics, tracing, or request logs. The hook must not
// block and must be safe for concurrent use.
type RequestHook func(ctx context.Context, method, url string)

// Option configures a Client. Options are validated by NewClient.
type Option func(*Client) error

// WithAPIKey sets the bearer token sent as Authorization: Bearer <key> on
// every request. Required for any tier above Free; Free is also key-gated
// but with a lower quota. An empty key is an error.
func WithAPIKey(key string) Option {
	return func(c *Client) error {
		key = strings.TrimSpace(key)
		if key == "" {
			return errors.New("cutline: API key must not be empty")
		}
		c.apiKey = key
		return nil
	}
}

// WithBaseURL points the client at another cutline-compatible server, such as
// a local gateway during development. The URL must be http or https.
func WithBaseURL(raw string) Option {
	return func(c *Client) error {
		trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
		u, err := url.Parse(trimmed)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errors.New("cutline: base URL must be an http(s) URL, got " + raw)
		}
		c.baseURL = trimmed
		return nil
	}
}

// WithHTTPClient replaces the HTTP client. Use it to add tracing, a custom
// transport, or a proxy. The client's own Timeout applies to each attempt.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) error {
		if h == nil {
			return errors.New("cutline: HTTP client must not be nil")
		}
		c.httpClient = h
		return nil
	}
}

// WithUserAgent sets the User-Agent header. Default is cutline-go-sdk/<Version>.
func WithUserAgent(ua string) Option {
	return func(c *Client) error {
		ua = strings.TrimSpace(ua)
		if ua == "" {
			return errors.New("cutline: user-agent must not be empty")
		}
		c.userAgent = ua
		return nil
	}
}

// WithRetry sets the number of retries and the base delay. maxRetries is
// clamped to MaxRetries; base is clamped to maxRetryBase. Use (0, 0) to
// disable retries entirely.
func WithRetry(maxRetries int, base time.Duration) Option {
	return func(c *Client) error {
		if maxRetries < 0 {
			return errors.New("cutline: maxRetries must not be negative")
		}
		if base < 0 {
			return errors.New("cutline: retry base must not be negative")
		}
		if maxRetries > MaxRetries {
			maxRetries = MaxRetries
		}
		if base > maxRetryBase {
			base = maxRetryBase
		}
		c.maxRetries = maxRetries
		c.retryBase = base
		return nil
	}
}

// WithRequestHook installs a hook called just before each HTTP attempt,
// including retries. See RequestHook.
func WithRequestHook(h RequestHook) Option {
	return func(c *Client) error {
		if h == nil {
			return errors.New("cutline: request hook must not be nil")
		}
		c.hook = h
		return nil
	}
}
