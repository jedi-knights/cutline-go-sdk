package cutline

import (
	"context"
	"net/url"
)

// Journey returns one person's career across clubs, schools, and levels,
// resolved by the public slug (e.g. "emma-smith-1a2b3c4d"). A merged person
// resolves to the survivor's journey; the response carries the survivor's
// canonical slug so callers can redirect.
//
// Returns IsNotFound-matching *APIError when no person carries that slug.
// Pro-tier feature; callers on lower tiers see ErrOutOfScope.
func (s *PeopleService) Journey(ctx context.Context, slug string) (*Journey, error) {
	if slug == "" {
		return nil, invalid("slug is required")
	}
	var out Journey
	op := operation{resource: "people", name: "Journey", route: "GET /v1/people/{slug}/journey"}
	if err := s.c.get(ctx, op, "/v1/people/"+url.PathEscape(slug)+"/journey", url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// JourneyByID returns the same journey as Journey, resolved by the opaque
// person_id (a UUID) that catalog projects onto roster rows. Use this from a
// roster row to link to a player's journey; the response carries the
// canonical slug for a redirect to /person/{slug}.
//
// Returns IsNotFound-matching *APIError when no person carries that id.
// Pro-tier feature; callers on lower tiers see ErrOutOfScope.
func (s *PeopleService) JourneyByID(ctx context.Context, id string) (*Journey, error) {
	if id == "" {
		return nil, invalid("id is required")
	}
	var out Journey
	op := operation{resource: "people", name: "JourneyByID", route: "GET /v1/people/by-id/{id}/journey"}
	if err := s.c.get(ctx, op, "/v1/people/by-id/"+url.PathEscape(id)+"/journey", url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
