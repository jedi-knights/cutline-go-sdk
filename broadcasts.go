package cutline

import (
	"context"
	"fmt"
	"net/url"
)

// Day returns per-game broadcast / streaming metadata for one calendar day.
// Callers merge the result with a Scoreboard response (same path shape, same
// sport/division/date) by game id to attach channel and URL to each game.
//
// Pro-tier feature; callers on lower tiers see ErrOutOfScope.
func (s *BroadcastsService) Day(ctx context.Context, p ScoreboardParams) (*Broadcasts, error) {
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
	path := fmt.Sprintf("/v1/broadcasts/%s/%s/%04d/%d/%d", p.Sport, p.Division, p.Year, p.Month, p.Day)
	var out Broadcasts
	op := operation{resource: "broadcasts", name: "Day", route: "GET /v1/broadcasts/{sport}/{division}/{year}/{month}/{day}"}
	if err := s.c.get(ctx, op, path, url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
