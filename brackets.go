package cutline

import (
	"context"
	"net/url"
)

// BracketParams identifies one sport+division bracket. Both fields are required.
type BracketParams struct {
	Sport    string
	Division string
}

// Tournament returns the NCAA national tournament bracket for the given
// sport+division, grouped into rounds. Before the real bracket is published,
// the response is a projection (Projected == true) so callers can render a
// bracket shape ahead of Selection Monday.
//
// Pro-tier feature; callers on lower tiers see ErrOutOfScope.
func (s *BracketsService) Tournament(ctx context.Context, p BracketParams) (*Tournament, error) {
	if err := validateBracketParams(p); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("sport", p.Sport)
	q.Set("division", p.Division)
	var out Tournament
	op := operation{resource: "brackets", name: "Tournament", route: "GET /v1/tournament"}
	if err := s.c.get(ctx, op, "/v1/tournament", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Projected returns a snapshot of the projected NCAA field — who would be
// selected if Selection Monday were today — along with the top 16 seed line
// and the bubble. Prediction, not reality; call Tournament for the real
// bracket once it has been published.
//
// Pro-tier feature; callers on lower tiers see ErrOutOfScope.
func (s *BracketsService) Projected(ctx context.Context, p BracketParams) (*ProjectedTournament, error) {
	if err := validateBracketParams(p); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("sport", p.Sport)
	q.Set("division", p.Division)
	var out ProjectedTournament
	op := operation{resource: "brackets", name: "Projected", route: "GET /v1/tournament/projected"}
	if err := s.c.get(ctx, op, "/v1/tournament/projected", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func validateBracketParams(p BracketParams) error {
	if p.Sport == "" {
		return invalid("sport is required")
	}
	if p.Division == "" {
		return invalid("division is required")
	}
	return nil
}
