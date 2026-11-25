package cutline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is The Cutline's public gateway.
	DefaultBaseURL = "https://thecutline.app"

	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 2
	defaultRetryBase  = 200 * time.Millisecond

	// maxBody caps how much of one response is read. Standings for the largest
	// division fits in a few hundred KiB; 8 MiB is a generous ceiling.
	maxBody = 8 << 20

	requestIDHeader = "x-request-id"
)

// Client talks to The Cutline. Create one with NewClient and reuse it; it is
// safe for concurrent use. Each resource is a field, so the method you want
// is one autocomplete away: client.Standings.Get, client.People.Journey.
type Client struct {
	// Standings, Scoreboard, Conferences, Teams, People, Brackets, and
	// Broadcasts are the API resources.
	Standings   *StandingsService
	Scoreboard  *ScoreboardService
	Conferences *ConferencesService
	Teams       *TeamsService
	Programs    *ProgramsService
	People      *PeopleService
	Brackets    *BracketsService
	Broadcasts  *BroadcastsService

	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string
	maxRetries int
	retryBase  time.Duration
	hook       RequestHook
}

// NewClient builds a Client. Options are applied in order and validated.
// The zero-option form works against DefaultBaseURL without authentication;
// use WithAPIKey for anything beyond the Free tier.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  "cutline-go-sdk/" + Version,
		maxRetries: defaultMaxRetries,
		retryBase:  defaultRetryBase,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	c.Standings = &StandingsService{c: c}
	c.Scoreboard = &ScoreboardService{c: c}
	c.Conferences = &ConferencesService{c: c}
	c.Teams = &TeamsService{c: c}
	c.Programs = &ProgramsService{c: c}
	c.People = &PeopleService{c: c}
	c.Brackets = &BracketsService{c: c}
	c.Broadcasts = &BroadcastsService{c: c}
	return c, nil
}

// operation names an SDK call for logging and future instrumentation.
type operation struct {
	// resource is the top-level noun ("standings", "scoreboard").
	resource string
	// name is the short method name ("Get", "List").
	name string
	// route is the low-cardinality route template ("GET /v1/standings").
	route string
}

// get executes a GET against the API and decodes JSON into out. It retries
// transient failures per Client's retry settings and honors ctx cancellation,
// including during backoff.
func (c *Client) get(ctx context.Context, op operation, path string, query url.Values, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return c.do(ctx, op, http.MethodGet, u, out)
}

func (c *Client) do(ctx context.Context, op operation, method, u string, out any) error {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			d := backoff(attempt-1, c.retryBase, 0)
			if err := sleep(ctx, d); err != nil {
				return err
			}
		}
		if c.hook != nil {
			c.hook(ctx, method, u)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, http.NoBody)
		if err != nil {
			return invalid("build request: %v", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("cutline: %s %s: %w", method, u, err)
			if isRetryableTransport(err) && attempt < c.maxRetries {
				continue
			}
			return lastErr
		}
		apiErr, retryAfter, done := c.handleResponse(resp, method, u, out, &lastErr)
		if done {
			return lastErr
		}
		if apiErr != nil && retryableStatus(apiErr.StatusCode) && attempt < c.maxRetries {
			if retryAfter > 0 {
				if err := sleep(ctx, capRetryAfter(retryAfter)); err != nil {
					return err
				}
			}
			lastErr = apiErr
			continue
		}
		if apiErr != nil {
			return apiErr
		}
		return nil
	}
	return lastErr
}

// handleResponse decodes a response into out or materializes the APIError.
// done is true when the loop should return immediately with lastErr set.
func (c *Client) handleResponse(resp *http.Response, method, u string, out any, lastErr *error) (apiErr *APIError, retryAfter time.Duration, done bool) {
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if readErr != nil {
		*lastErr = fmt.Errorf("cutline: read body: %w", readErr)
		return nil, 0, true
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Clear any error from a prior retry attempt; this one succeeded.
		*lastErr = nil
		if out == nil {
			return nil, 0, true
		}
		if err := json.Unmarshal(body, out); err != nil {
			*lastErr = invalidResponse("%s %s: %v", method, u, err)
			return nil, 0, true
		}
		return nil, 0, true
	}
	apiErr = &APIError{
		StatusCode: resp.StatusCode,
		Message:    http.StatusText(resp.StatusCode),
		RequestID:  resp.Header.Get(requestIDHeader),
		Method:     method,
		URL:        u,
	}
	apiErr.decorateFromBody(body)
	retryAfter = parseRetryAfter(resp.Header)
	return apiErr, retryAfter, false
}

// isRetryableTransport classifies a transport-level error. Timeouts and DNS
// temporary failures are retried; anything else surfaces immediately.
func isRetryableTransport(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	return false
}
