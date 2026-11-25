# cutline-go-sdk

Go SDK for The Cutline (thecutline.app). A read-only client for NCAA standings, scoreboards, conferences, player career journeys, and conference tournament brackets. Mirrors the shape of `jsonplaceholder-sdk`; hexagonal in spirit but a flat library package (no cmd/ or internal/ split — this is a library, not a service).

## Layout

One package, `cutline`, at the repo root. One file per responsibility:

- `client.go` — Client struct, NewClient, the transport + retry loop (`get`, `do`, `handleResponse`).
- `options.go` — `Option` and the `With*` constructors.
- `errors.go` — `*APIError`, sentinels (`ErrInvalidArgument`, `ErrInvalidResponse`), predicates (`IsNotFound`, `IsPaymentRequired`, `IsOutOfScope`, `IsRateLimited`), `codeForStatus`, `invalid` / `invalidResponse` helpers.
- `retry.go` — `backoff`, `sleep`, `parseRetryAfter`, `retryableStatus`, `capRetryAfter`, named bounds (`maxBackoff`, `retryAfterCap`).
- `services.go` — all 8 service struct declarations in one place (declarations only, methods live in per-resource files).
- `models.go` — every response/sub-response type (one place for the full data surface).
- `standings.go` / `scoreboard.go` / `conferences.go` / `teams.go` / `programs.go` / `people.go` / `brackets.go` / `broadcasts.go` — one file per resource, each with its service methods.
- `doc.go` — package-level documentation with runnable examples.
- `version.go` — `Version` constant, build-time injectable via `-ldflags`.
- `examples/` — runnable programs against the live service (`make examples`).

## Conventions

- Every request method takes `ctx context.Context` first.
- Every loop has a named bound (`MaxRetries`, `maxBackoff`, `retryAfterCap`).
- `*APIError.Unwrap` exposes a go-platform `*apperrors.AppError` (code derived from HTTP status); every sentinel error is an `*apperrors.AppError` too.
- Failures are typed: `*APIError` (inspect with `errors.As`), `ErrInvalidArgument` (checked before any request), `ErrInvalidResponse` (200 with un-parseable body, not retried).
- Validate caller arguments in the service method **before** calling `c.get` — empty required strings and out-of-range integers return an `ErrInvalidArgument`-wrapping error. The server-side validation is the fallback, not the primary.
- Zero `ScoreboardParams` is invalid (every field is required). Zero `TeamListParams` is invalid for the same reason. Document explicit "optional" filters as pointer or empty-string.
- The service is read-only. No write methods, ever.
- Public API changes follow semver and need a `CHANGELOG.md` entry. A new resource, method, option, or response field is additive; renaming anything exported is breaking.

## Dependencies — strict

Library code may import:

- `github.com/jedi-knights/go-platform/apperrors` (zero-dependency)
- `go.opentelemetry.io/otel` + `otel/metric` + `otel/trace` + `otel/propagation` (API only; **never** the OTel SDK, exporters, gRPC, or Prometheus)
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`
- Go stdlib

Library code **may not** import:

- `github.com/jedi-knights/go-platform/otel` (the application bootstrap pulls ~240 packages)
- `github.com/jedi-knights/go-platform/httpserver` or `httpmw` (server-side)
- Any OpenTelemetry exporter, SDK, or transport
- Any log library beyond `log/slog`

Test-only code (`_test.go`) may import the OTel SDK and test utilities.

## Telemetry (planned, API only)

Telemetry is not yet wired. When it lands it will follow the `jsonplaceholder-sdk` pattern:

- Spans: `cutline.<resource>.<operation>` (e.g. `cutline.standings.Get`).
- Metrics: `cutline.client.operation_duration` (histogram, seconds) + `cutline.client.retries` (counter).
- Attribute keys mirror go-platform's `resource.*` vocabulary (ADR-0019). Keys are copied as strings — never import the attrs package because it drags the OTel SDK transitively.
- Turnkey on-off: on by default; `WithoutTelemetry()` disables it; no-op until the host app installs OTel providers.

The current `WithRequestHook` is the manual seam; it suffices for request logs and custom metrics in the meantime.

## Adding a resource or method

1. Add the response types to `models.go`.
2. Add the service struct declaration to `services.go` (if it's a new resource).
3. Add the field to `client.go`'s `Client` struct and wire it in `NewClient`.
4. Write a new `<resource>.go` with the methods. Each method:
   - Takes `ctx` first.
   - Validates caller input (`invalid("…")`) before hitting the network.
   - Builds an `operation{resource, name, route}`.
   - Calls `c.get(ctx, op, path, query, &out)`.
   - Returns `*T, error`.
5. Update the `Resources` table in `README.md` and the overview in `doc.go`.
6. Bump the semver and add a `CHANGELOG.md` entry.

## Commands

```
make test     # go test -race ./...
make lint     # golangci-lint run ./...
make vuln     # govulncheck ./...
make examples # go run each example (hits the live service)
```

## Known drift / TODO

- Telemetry wiring (spans + metrics via OTel API).
- Pagination helpers (`Page[T]`, `iter.Seq2`) — cutline's current endpoints are not paginated, but the pattern is ready to drop in from `jsonplaceholder-sdk` when needed.
- `WithoutTelemetry` option — stubbed for parity with the reference but currently unused.
- Pro-soccer per-team endpoints (`/v1/teams/{sport}/{division}/{slug}/games` and `.../roster`) are Team-tier and intentionally not yet in the SDK.
- `/v1/divisions`, `/v1/seasons`, `/v1/competition/{id}`, `/v1/club/{id}` are small catalog helpers — add when a consumer asks.
