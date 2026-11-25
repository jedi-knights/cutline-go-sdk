# cutline-go-sdk

A Go SDK for [The Cutline](https://thecutline.app), a read-only API over NCAA standings, scoreboards, conferences, player career journeys, and conference tournament brackets. Instrumented for OpenTelemetry via the API package (never the SDK, exporters, or gRPC), with a dependency surface limited to `apperrors` from `go-platform` plus the OpenTelemetry API.

The code is small enough to read in one sitting: one client, one options file, one resource file per API noun.

## Install

Requires Go 1.23 or later.

```bash
go get github.com/jedi-knights/cutline-go-sdk
```

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    cutline "github.com/jedi-knights/cutline-go-sdk"
)

func main() {
    client, err := cutline.NewClient(cutline.WithAPIKey(os.Getenv("CUTLINE_API_KEY")))
    if err != nil {
        log.Fatal(err)
    }

    standings, err := client.Standings.Get(context.Background(), cutline.StandingsParams{
        Sport:    "soccer-women",
        Division: "d1",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d teams; leader: %s RPI %.4f\n",
        standings.Count,
        standings.Standings[0].TeamSlug,
        standings.Standings[0].RPI)
}
```

Run the included example against the live service with `make examples` (sets `CUTLINE_API_KEY` from your environment) or read the code in [`examples/basic`](examples/basic/main.go).

## Resources

Each resource is a field on the Client. Methods take `context.Context` first and return typed responses or typed errors.

| Resource | Methods | Tier gating |
|---|---|---|
| [`client.Standings`](standings.go) | `Get(ctx, StandingsParams)` | Free |
| [`client.Scoreboard`](scoreboard.go) | `Day(ctx, ScoreboardParams)` | Free |
| [`client.Conferences`](conferences.go) | `List(ctx, ConferenceListParams)`, `Get(ctx, slug)`, `Tournament(ctx, slug)` | List/Get: Free; Tournament: Pro |
| [`client.Teams`](teams.go) | `List(ctx, TeamListParams)`, `Get(ctx, slug)` | List: Free; Get: Developer |
| [`client.Programs`](programs.go) | `Get(ctx, slug)`, `Roster(ctx, RosterParams)` | Get: Developer; Roster: Team |
| [`client.People`](people.go) | `Journey(ctx, slug)`, `JourneyByID(ctx, id)` | Pro |
| [`client.Brackets`](brackets.go) | `Tournament(ctx, BracketParams)`, `Projected(ctx, BracketParams)` | Pro |
| [`client.Broadcasts`](broadcasts.go) | `Day(ctx, ScoreboardParams)` | Pro |

Calling a method outside your plan returns an `*APIError` with status 403 — check it with `cutline.IsOutOfScope(err)`.

The Cutline is read-only, so this SDK is read-only.

## Authentication

Every tier above Free requires an API key passed as `Authorization: Bearer <key>`. Get one at [thecutline.app/account/api-keys](https://thecutline.app/account/api-keys) (surface in flight).

```go
client, _ := cutline.NewClient(cutline.WithAPIKey("sk_live_…"))
```

The Free tier also accepts a key (which it uses for quota accounting) but tolerates the zero-option form for the public D1 women's soccer endpoints.

## Errors

Every failure is typed. Use the predicates for the common cases and `errors.As` on `*APIError` for the rest.

```go
standings, err := client.Standings.Get(ctx, params)
switch {
case cutline.IsPaymentRequired(err):
    // Plan lapsed; prompt renewal.
case cutline.IsOutOfScope(err):
    // Sport/division not in this plan; prompt upgrade.
case cutline.IsRateLimited(err):
    // Over per-minute limit or Free monthly quota; back off.
case cutline.IsNotFound(err):
    // No data for that sport/division.
case errors.Is(err, cutline.ErrInvalidArgument):
    // Caller sent an invalid argument; do not retry.
}

var apiErr *cutline.APIError
if errors.As(err, &apiErr) {
    log.Printf("cutline %d %s: %s (request %s)",
        apiErr.StatusCode, apiErr.Code, apiErr.Detail, apiErr.RequestID)
}
```

### Interop with `apperrors`

Every SDK error chain unwraps to a `*apperrors.AppError` from go-platform, so code already built on go-platform works without extra translation:

```go
import "github.com/jedi-knights/go-platform/apperrors"

if apperrors.IsNotFound(err) { ... }
```

The `AppError` code is derived from the HTTP status (404 → `NOT_FOUND`, 429 → `RATE_LIMIT`, 403 → `FORBIDDEN`, and so on).

## Retries and transport

The SDK retries transient failures automatically: timeouts, `500`, `502`, `503`, `504`. `402`, `403`, `404`, and `429` are not retried — they describe plan or scope state that will not change on an immediate retry.

Backoff is exponential from a configurable base, with jitter in `[d/2, d]`. A server-sent `Retry-After` raises the delay (capped at 30s).

Tune with `WithRetry`:

```go
client, _ := cutline.NewClient(
    cutline.WithAPIKey(key),
    cutline.WithRetry(3, 250*time.Millisecond),
)
```

`WithHTTPClient` takes a `*http.Client` and lets you plug in tracing, a custom transport, or a proxy. The `WithRequestHook` option installs a hook called before every attempt (retries included) — useful for request logs and custom metrics.

## Configuration

| Option | Default | Purpose |
|---|---|---|
| `WithAPIKey(key)` | unset | Bearer token for authenticated tiers |
| `WithBaseURL(url)` | `https://thecutline.app` | Point at a dev server or a VCR fixture |
| `WithHTTPClient(hc)` | 30s timeout | Replace the underlying `*http.Client` |
| `WithUserAgent(ua)` | `cutline-go-sdk/<version>` | Identify the integration in logs |
| `WithRetry(n, base)` | 2 retries, 200ms base | Tune the retry loop |
| `WithRequestHook(fn)` | nil | Observe every attempt |

## Design

See the SDK design series on omarcrosby.com; the short form:

- Dependencies are deliberately minimal. Library code imports `apperrors` and the OpenTelemetry API only — never the OTel SDK, exporters, gRPC, or Prometheus. Never `go-platform/otel` (the application bootstrap is about 240 packages).
- Every request method takes `ctx` first. Every loop has a named bound (`MaxRetries`).
- Public API changes follow semver; breaking renames need a `CHANGELOG.md` entry.
- The service is read-only, so there are no write methods.

## Development

```bash
make test       # go test ./...
make lint       # golangci-lint run ./...
make vuln       # govulncheck ./...
make examples   # hits the live service — needs CUTLINE_API_KEY
```

## License

MIT. See [LICENSE](LICENSE).
