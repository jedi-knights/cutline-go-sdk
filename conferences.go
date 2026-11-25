package cutline

import (
	"context"
	"net/url"
)

// ConferenceListParams filters a conference listing. All fields are optional;
// the zero value lists every conference in the catalog.
type ConferenceListParams struct {
	// Sport filters to conferences that field teams in this sport.
	Sport string
	// Division filters to conferences that field teams at this division.
	Division string
}

// List returns the conference registry, optionally filtered by sport and
// division. Returns an error wrapping ErrInvalidArgument for malformed
// filters and an *APIError on upstream failure.
func (s *ConferencesService) List(ctx context.Context, p ConferenceListParams) (*ConferencesList, error) {
	q := url.Values{}
	if p.Sport != "" {
		q.Set("sport", p.Sport)
	}
	if p.Division != "" {
		q.Set("division", p.Division)
	}
	var out ConferencesList
	op := operation{resource: "conferences", name: "List", route: "GET /v1/conferences"}
	if err := s.c.get(ctx, op, "/v1/conferences", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns one conference by slug (e.g. "acc", "sec", "big-ten"). Returns
// IsNotFound-matching *APIError when no such conference exists.
func (s *ConferencesService) Get(ctx context.Context, slug string) (*Conference, error) {
	if slug == "" {
		return nil, invalid("slug is required")
	}
	var out Conference
	op := operation{resource: "conferences", name: "Get", route: "GET /v1/conferences/{slug}"}
	if err := s.c.get(ctx, op, "/v1/conferences/"+url.PathEscape(slug), url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Tournament returns the games a conference has played (or scheduled) in its
// post-regular-season tournament window. Pro tier and above.
//
// Returns an empty Games list — not an error — when the conference has no
// regular_season_end configured. This is the signal callers use to fall back
// to a projected bracket.
//
// Pro-tier feature; callers on lower tiers see ErrOutOfScope.
func (s *ConferencesService) Tournament(ctx context.Context, slug string) (*ConferenceTournament, error) {
	if slug == "" {
		return nil, invalid("slug is required")
	}
	var out ConferenceTournament
	op := operation{resource: "conferences", name: "Tournament", route: "GET /v1/conferences/{slug}/tournament"}
	if err := s.c.get(ctx, op, "/v1/conferences/"+url.PathEscape(slug)+"/tournament", url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
