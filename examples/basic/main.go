// Command basic fetches today's D1 women's soccer standings and prints the
// top five teams by RPI. It also demonstrates the typed error predicates.
//
// Usage:
//
//	CUTLINE_API_KEY=sk_live_… go run ./examples/basic
//
// The CUTLINE_API_KEY env var is required for every tier above Free. The Free
// tier also accepts a key (and uses it for quota accounting) but tolerates an
// empty key against the public D1 women's soccer standings endpoint.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	cutline "github.com/jedi-knights/cutline-go-sdk"
)

const (
	topN     = 5
	sport    = "soccer-women"
	division = "d1"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	opts := []cutline.Option{}
	if key := os.Getenv("CUTLINE_API_KEY"); key != "" {
		opts = append(opts, cutline.WithAPIKey(key))
	}
	client, err := cutline.NewClient(opts...)
	if err != nil {
		return err
	}

	standings, err := client.Standings.Get(ctx, cutline.StandingsParams{
		Sport:    sport,
		Division: division,
	})
	if err != nil {
		return describeError(err)
	}

	fmt.Printf("%s %s — %d teams\n", standings.Sport, standings.Division, standings.Count)
	for i, row := range standings.Standings {
		if i >= topN {
			break
		}
		fmt.Printf("  %2d. %-24s RPI %.4f  (%d-%d-%d)\n",
			i+1, row.TeamSlug, row.RPI, row.Wins, row.Losses, row.Ties)
	}
	return nil
}

// describeError translates typed errors into human messages. In application
// code the caller would branch on these predicates to decide what to do
// (upgrade prompt, retry, surface a form error, etc.).
func describeError(err error) error {
	switch {
	case cutline.IsPaymentRequired(err):
		return fmt.Errorf("plan expired or no active subscription: %w", err)
	case cutline.IsOutOfScope(err):
		return fmt.Errorf("the requested sport+division is not in your plan: %w", err)
	case cutline.IsRateLimited(err):
		return fmt.Errorf("rate limited: slow down or upgrade: %w", err)
	case cutline.IsNotFound(err):
		return fmt.Errorf("no data for that sport+division: %w", err)
	case errors.Is(err, cutline.ErrInvalidArgument):
		return fmt.Errorf("invalid argument: %w", err)
	}
	var apiErr *cutline.APIError
	if errors.As(err, &apiErr) {
		log.Printf("api error: status=%d code=%s detail=%q request_id=%s",
			apiErr.StatusCode, apiErr.Code, apiErr.Detail, apiErr.RequestID)
	}
	return err
}
