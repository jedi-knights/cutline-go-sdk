package cutline

import (
	"context"
	"fmt"
	"net/url"
)

// ScoreboardParams identifies one calendar day of scoreboard data. All four
// fields are required.
type ScoreboardParams struct {
	Sport    string // e.g. "soccer-women"
	Division string // e.g. "d1"
	Year     int    // 4-digit calendar year
	Month    int    // 1-12
	Day      int    // 1-31
}

// Day returns the scoreboard for the given sport, division, and calendar day.
// Returns an error wrapping ErrInvalidArgument for empty sport/division or
// out-of-range year/month/day, an *APIError (404) if the route has no data,
// and ErrPaymentRequired / ErrOutOfScope per the caller's plan.
func (s *ScoreboardService) Day(ctx context.Context, p ScoreboardParams) (*Scoreboard, error) {
	if p.Sport == "" {
		return nil, invalid("sport is required")
	}
	if p.Division == "" {
		return nil, invalid("division is required")
	}
	if p.Year < 1900 || p.Year > 9999 {
		return nil, invalid("year must be a four-digit calendar year, got %d", p.Year)
	}
	if p.Month < 1 || p.Month > 12 {
		return nil, invalid("month must be 1-12, got %d", p.Month)
	}
	if p.Day < 1 || p.Day > 31 {
		return nil, invalid("day must be 1-31, got %d", p.Day)
	}
	path := fmt.Sprintf("/v1/scoreboard/%s/%s/%04d/%d/%d", p.Sport, p.Division, p.Year, p.Month, p.Day)
	var out Scoreboard
	op := operation{resource: "scoreboard", name: "Day", route: "GET /v1/scoreboard/{sport}/{division}/{year}/{month}/{day}"}
	if err := s.c.get(ctx, op, path, url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
