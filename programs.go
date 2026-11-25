package cutline

import (
	"context"
	"net/url"
)

// Get returns one school-level program by slug (e.g. "north-carolina").
// Returns IsNotFound-matching *APIError when no program carries that slug.
func (s *ProgramsService) Get(ctx context.Context, slug string) (*Program, error) {
	if slug == "" {
		return nil, invalid("slug is required")
	}
	var out Program
	op := operation{resource: "programs", name: "Get", route: "GET /v1/programs/{slug}"}
	if err := s.c.get(ctx, op, "/v1/programs/"+url.PathEscape(slug), url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RosterParams narrows a roster lookup. Slug is required; Season is optional
// (empty string = current season).
type RosterParams struct {
	Slug   string
	Season string
}

// Roster returns the program's roster for one season. When Season is empty,
// the current season is returned.
//
// Team-tier feature; callers on lower tiers see ErrOutOfScope when the
// endpoint isn't in their plan.
func (s *ProgramsService) Roster(ctx context.Context, p RosterParams) (*Roster, error) {
	if p.Slug == "" {
		return nil, invalid("slug is required")
	}
	q := url.Values{}
	if p.Season != "" {
		q.Set("season", p.Season)
	}
	var out Roster
	op := operation{resource: "programs", name: "Roster", route: "GET /v1/programs/{slug}/roster"}
	if err := s.c.get(ctx, op, "/v1/programs/"+url.PathEscape(p.Slug)+"/roster", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
