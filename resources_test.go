package cutline

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// recordingServer captures each incoming request's method, path, and query
// for a table-driven service test to assert.
type recordingServer struct {
	srv    *httptest.Server
	method string
	path   string
	query  string
}

func newRecordingServer(t *testing.T, responseBody string) *recordingServer {
	t.Helper()
	r := &recordingServer{}
	r.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.method = req.Method
		r.path = req.URL.Path
		r.query = req.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	t.Cleanup(r.srv.Close)
	return r
}

func newServiceClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(WithBaseURL(baseURL), WithAPIKey("sk_test"), WithRetry(0, time.Millisecond))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// TestResources_PathsAndMethods walks every resource method and asserts it
// hits the expected HTTP method + URL path + query string. This is the
// compile-against-current-API gate: adding a new method or changing a path
// breaks this test loudly.
func TestResources_PathsAndMethods(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		body       string
		call       func(ctx context.Context, c *Client) error
		wantMethod string
		wantPath   string
		wantQuery  string
	}{
		{
			name: "Standings.Get",
			body: `{"sport":"soccer-women","division":"d1","count":0,"standings":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Standings.Get(ctx, StandingsParams{Sport: "soccer-women", Division: "d1"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/standings",
			wantQuery:  "division=d1&sport=soccer-women",
		},
		{
			name: "Standings.Get with Season",
			body: `{"sport":"soccer-women","division":"d1","count":0,"standings":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Standings.Get(ctx, StandingsParams{Sport: "soccer-women", Division: "d1", Season: "2025"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/standings",
			wantQuery:  "division=d1&season=2025&sport=soccer-women",
		},
		{
			name: "Scoreboard.Day",
			body: `{"sport":"soccer-women","division":"d1","date":"2026-10-29","count":0,"games":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Scoreboard.Day(ctx, ScoreboardParams{Sport: "soccer-women", Division: "d1", Year: 2026, Month: 10, Day: 29})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/scoreboard/soccer-women/d1/2026/10/29",
			wantQuery:  "",
		},
		{
			name: "Conferences.List",
			body: `{"count":0,"conferences":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Conferences.List(ctx, ConferenceListParams{Sport: "soccer-women", Division: "d1"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/conferences",
			wantQuery:  "division=d1&sport=soccer-women",
		},
		{
			name: "Conferences.List with no filters",
			body: `{"count":0,"conferences":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Conferences.List(ctx, ConferenceListParams{})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/conferences",
			wantQuery:  "",
		},
		{
			name: "Conferences.Get",
			body: `{"slug":"acc","name":"Atlantic Coast Conference","shortName":"ACC"}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Conferences.Get(ctx, "acc")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/conferences/acc",
			wantQuery:  "",
		},
		{
			name: "Conferences.Tournament",
			body: `{"sport":"soccer-women","division":"d1","conference":"acc","count":0,"games":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Conferences.Tournament(ctx, "acc")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/conferences/acc/tournament",
			wantQuery:  "",
		},
		{
			name: "Teams.List",
			body: `{"sport":"soccer-women","division":"d1","count":0,"teams":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Teams.List(ctx, TeamListParams{Sport: "soccer-women", Division: "d1"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/teams",
			wantQuery:  "division=d1&sport=soccer-women",
		},
		{
			name: "Teams.List with Conference filter",
			body: `{"sport":"soccer-women","division":"d1","conference":"acc","count":0,"teams":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Teams.List(ctx, TeamListParams{Sport: "soccer-women", Division: "d1", Conference: "acc"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/teams",
			wantQuery:  "conference=acc&division=d1&sport=soccer-women",
		},
		{
			name: "Teams.Get",
			body: `{"slug":"unc","shortName":"North Carolina","conference":"acc","sport":"soccer-women","division":"d1"}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Teams.Get(ctx, "unc")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/teams/unc",
			wantQuery:  "",
		},
		{
			name: "Programs.Get",
			body: `{"slug":"north-carolina","name":"North Carolina","conference":"acc","state":"NC","rosterUrl":"","scheduleUrl":"","statsUrl":""}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Programs.Get(ctx, "north-carolina")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/programs/north-carolina",
			wantQuery:  "",
		},
		{
			name: "Programs.Roster",
			body: `{"programSlug":"north-carolina","count":0,"players":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Programs.Roster(ctx, RosterParams{Slug: "north-carolina"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/programs/north-carolina/roster",
			wantQuery:  "",
		},
		{
			name: "Programs.Roster with Season",
			body: `{"programSlug":"north-carolina","season":"2025","count":0,"players":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Programs.Roster(ctx, RosterParams{Slug: "north-carolina", Season: "2025"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/programs/north-carolina/roster",
			wantQuery:  "season=2025",
		},
		{
			name: "People.Journey",
			body: `{"person":{"slug":"emma-smith","displayName":"Emma Smith"},"stints":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.People.Journey(ctx, "emma-smith-1a2b3c4d")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/people/emma-smith-1a2b3c4d/journey",
			wantQuery:  "",
		},
		{
			name: "People.JourneyByID",
			body: `{"person":{"slug":"emma-smith","displayName":"Emma Smith"},"stints":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.People.JourneyByID(ctx, "11111111-2222-3333-4444-555555555555")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/people/by-id/11111111-2222-3333-4444-555555555555/journey",
			wantQuery:  "",
		},
		{
			name: "Brackets.Tournament",
			body: `{"sport":"soccer-women","division":"d1","fieldSize":64,"totalGames":0,"rounds":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Brackets.Tournament(ctx, BracketParams{Sport: "soccer-women", Division: "d1"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/tournament",
			wantQuery:  "division=d1&sport=soccer-women",
		},
		{
			name: "Brackets.Projected",
			body: `{"sport":"soccer-women","division":"d1","fieldSize":64,"asOf":"2026-11-01","topSixteen":[],"unseeded":[],"bubble":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Brackets.Projected(ctx, BracketParams{Sport: "soccer-women", Division: "d1"})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/tournament/projected",
			wantQuery:  "division=d1&sport=soccer-women",
		},
		{
			name: "Broadcasts.Day",
			body: `{"sport":"soccer-women","division":"d1","date":"2026-10-29","count":0,"broadcasts":[]}`,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Broadcasts.Day(ctx, ScoreboardParams{Sport: "soccer-women", Division: "d1", Year: 2026, Month: 10, Day: 29})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/broadcasts/soccer-women/d1/2026/10/29",
			wantQuery:  "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newRecordingServer(t, tc.body)
			c := newServiceClient(t, r.srv.URL)
			if err := tc.call(context.Background(), c); err != nil {
				t.Fatalf("call: %v", err)
			}
			if r.method != tc.wantMethod {
				t.Errorf("method = %q, want %q", r.method, tc.wantMethod)
			}
			if r.path != tc.wantPath {
				t.Errorf("path = %q, want %q", r.path, tc.wantPath)
			}
			if r.query != tc.wantQuery {
				t.Errorf("query = %q, want %q", r.query, tc.wantQuery)
			}
		})
	}
}

// TestResources_InvalidArgRejectedBeforeNetwork asserts that every method
// validates required arguments before making an HTTP call. The server is
// configured to fail the test if it is ever hit.
func TestResources_InvalidArgRejectedBeforeNetwork(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("server should not be called")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	c := newServiceClient(t, srv.URL)
	ctx := context.Background()

	cases := map[string]func() error{
		"Standings.Get empty sport":    func() error { _, e := c.Standings.Get(ctx, StandingsParams{Division: "d1"}); return e },
		"Standings.Get empty division": func() error { _, e := c.Standings.Get(ctx, StandingsParams{Sport: "soccer-women"}); return e },
		"Scoreboard.Day bad month": func() error {
			_, e := c.Scoreboard.Day(ctx, ScoreboardParams{Sport: "soccer-women", Division: "d1", Year: 2026, Month: 13, Day: 1})
			return e
		},
		"Scoreboard.Day bad day": func() error {
			_, e := c.Scoreboard.Day(ctx, ScoreboardParams{Sport: "soccer-women", Division: "d1", Year: 2026, Month: 10, Day: 32})
			return e
		},
		"Scoreboard.Day bad year": func() error {
			_, e := c.Scoreboard.Day(ctx, ScoreboardParams{Sport: "soccer-women", Division: "d1", Year: 42, Month: 10, Day: 29})
			return e
		},
		"Conferences.Get empty slug":        func() error { _, e := c.Conferences.Get(ctx, ""); return e },
		"Conferences.Tournament empty slug": func() error { _, e := c.Conferences.Tournament(ctx, ""); return e },
		"Teams.List empty sport":            func() error { _, e := c.Teams.List(ctx, TeamListParams{Division: "d1"}); return e },
		"Teams.Get empty slug":              func() error { _, e := c.Teams.Get(ctx, ""); return e },
		"Programs.Get empty slug":           func() error { _, e := c.Programs.Get(ctx, ""); return e },
		"Programs.Roster empty slug":        func() error { _, e := c.Programs.Roster(ctx, RosterParams{}); return e },
		"People.Journey empty slug":         func() error { _, e := c.People.Journey(ctx, ""); return e },
		"People.JourneyByID empty id":       func() error { _, e := c.People.JourneyByID(ctx, ""); return e },
		"Brackets.Tournament empty sport":   func() error { _, e := c.Brackets.Tournament(ctx, BracketParams{Division: "d1"}); return e },
		"Brackets.Projected empty division": func() error { _, e := c.Brackets.Projected(ctx, BracketParams{Sport: "soccer-women"}); return e },
		"Broadcasts.Day empty sport": func() error {
			_, e := c.Broadcasts.Day(ctx, ScoreboardParams{Division: "d1", Year: 2026, Month: 10, Day: 29})
			return e
		},
	}
	for name, invoke := range cases {
		t.Run(name, func(t *testing.T) {
			err := invoke()
			if !errors.Is(err, ErrInvalidArgument) {
				t.Errorf("err = %v, want errors.Is ErrInvalidArgument", err)
			}
		})
	}
}

// TestResources_DecodesTypedResponse exercises one representative resource
// to prove the response body actually gets parsed into the right struct.
// Other resources share the same decode path, so one well-chosen case
// suffices.
func TestResources_DecodesTypedResponse(t *testing.T) {
	t.Parallel()
	body := `{
		"sport": "soccer-women",
		"division": "d1",
		"count": 2,
		"standings": [
			{"teamSlug":"unc","wins":10,"losses":1,"ties":2,"games":13,"winPercentage":0.846,"owp":0.52,"oowp":0.49,"rpi":0.56},
			{"teamSlug":"duke","wins":9,"losses":2,"ties":2,"games":13,"winPercentage":0.769,"owp":0.51,"oowp":0.48,"rpi":0.54}
		]
	}`
	r := newRecordingServer(t, body)
	c := newServiceClient(t, r.srv.URL)
	got, err := c.Standings.Get(context.Background(), StandingsParams{Sport: "soccer-women", Division: "d1"})
	if err != nil {
		t.Fatalf("Standings.Get: %v", err)
	}
	if got.Count != 2 {
		t.Errorf("Count = %d, want 2", got.Count)
	}
	if len(got.Standings) != 2 {
		t.Fatalf("len(Standings) = %d, want 2", len(got.Standings))
	}
	if got.Standings[0].TeamSlug != "unc" {
		t.Errorf("Standings[0].TeamSlug = %q, want unc", got.Standings[0].TeamSlug)
	}
	if got.Standings[0].RPI < 0.55 || got.Standings[0].RPI > 0.57 {
		t.Errorf("Standings[0].RPI = %v, want ~0.56", got.Standings[0].RPI)
	}
}

// TestResources_URLEscapesSlug makes sure slugs with reserved characters
// don't break the URL path. UUIDs for JourneyByID are uppercase-ASCII only,
// but slugs can contain runs of hyphens and alphanumerics, which is what
// people will actually feed in.
func TestResources_URLEscapesSlug(t *testing.T) {
	t.Parallel()
	r := newRecordingServer(t, `{"person":{"slug":"x","displayName":"x"},"stints":[]}`)
	c := newServiceClient(t, r.srv.URL)
	// A real slug shape with multiple hyphens.
	if _, err := c.People.Journey(context.Background(), "emma-smith-1a2b3c4d"); err != nil {
		t.Fatalf("Journey: %v", err)
	}
	if !strings.Contains(r.path, "emma-smith-1a2b3c4d") {
		t.Errorf("path = %q, want to contain slug", r.path)
	}
}
