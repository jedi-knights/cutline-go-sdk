// Package cutline is a Go SDK for The Cutline (thecutline.app), a read-only
// API over NCAA standings, scoreboards, conferences, player career journeys,
// and conference tournament brackets.
//
// Create a Client once and reuse it. Resources are fields with Get and List
// methods:
//
//	client, err := cutline.NewClient(cutline.WithAPIKey("sk_live_…"))
//	if err != nil {
//		return err
//	}
//
//	standings, err := client.Standings.Get(ctx, cutline.StandingsParams{
//		Sport:    "soccer-women",
//		Division: "d1",
//	})
//
//	journey, err := client.People.Journey(ctx, "sophia-smith")
//
// The SDK owns the cross-cutting work so callers do not repeat it: transient
// failures are retried with jittered exponential backoff honoring Retry-After,
// authentication is a single Option, and failures are typed values (APIError,
// ErrOutOfScope, ErrPaymentRequired, ErrRateLimited, ErrInvalidArgument)
// rather than strings. Every method takes a context.Context first and honors
// cancellation, including during backoff.
//
// Errors carry a go-platform *apperrors.AppError in their chain, so
// apperrors.IsNotFound(err) works on them.
//
// The Cutline API is read-only, so this SDK is read-only too.
package cutline
