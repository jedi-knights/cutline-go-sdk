package cutline

import (
	"context"
	"net/url"
)

// TeamListParams filters a team listing. Sport and Division are required to
// pin the scope; Conference is optional.
type TeamListParams struct {
	Sport      string
	Division   string
	Conference string // optional slug filter, e.g. "acc"
}

// List returns the teams registered for a given sport+division, optionally
// narrowed to one conference. Returns an error wrapping ErrInvalidArgument
// for empty sport/division and an *APIError on upstream failure.
func (s *TeamsService) List(ctx context.Context, p TeamListParams) (*TeamsList, error) {
	if p.Sport == "" {
		return nil, invalid("sport is required")
	}
	if p.Division == "" {
		return nil, invalid("division is required")
	}
	q := url.Values{}
	q.Set("sport", p.Sport)
	q.Set("division", p.Division)
	if p.Conference != "" {
		q.Set("conference", p.Conference)
	}
	var out TeamsList
	op := operation{resource: "teams", name: "List", route: "GET /v1/teams"}
	if err := s.c.get(ctx, op, "/v1/teams", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns one team by slug (e.g. "unc", "duke"). Returns IsNotFound-
// matching *APIError when no team carries that slug.
func (s *TeamsService) Get(ctx context.Context, slug string) (*Team, error) {
	if slug == "" {
		return nil, invalid("slug is required")
	}
	var out Team
	op := operation{resource: "teams", name: "Get", route: "GET /v1/teams/{slug}"}
	if err := s.c.get(ctx, op, "/v1/teams/"+url.PathEscape(slug), url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
