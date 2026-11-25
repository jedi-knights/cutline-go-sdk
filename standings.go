package cutline

import (
	"context"
	"net/url"
)

// StandingsParams filters a standings query. Sport and Division are required.
// A non-empty Season recomputes the table from that season's own games and
// conferences; omit it (empty string) for the current season.
type StandingsParams struct {
	// Sport is a cutline SportID, e.g. "soccer-women", "soccer-men".
	Sport string
	// Division is a cutline Division, e.g. "d1", "d2".
	Division string
	// Season is optional. The empty string means "current season".
	Season string
}

// Get returns the RPI standings for the given sport+division (and optionally
// season). Returns an error wrapping ErrInvalidArgument for empty sport or
// division, *APIError on upstream failure, ErrPaymentRequired if the caller's
// plan has lapsed, ErrOutOfScope if the sport or division is not in the
// caller's plan.
func (s *StandingsService) Get(ctx context.Context, p StandingsParams) (*Standings, error) {
	if p.Sport == "" {
		return nil, invalid("sport is required")
	}
	if p.Division == "" {
		return nil, invalid("division is required")
	}
	q := url.Values{}
	q.Set("sport", p.Sport)
	q.Set("division", p.Division)
	if p.Season != "" {
		q.Set("season", p.Season)
	}
	var out Standings
	op := operation{resource: "standings", name: "Get", route: "GET /v1/standings"}
	if err := s.c.get(ctx, op, "/v1/standings", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
