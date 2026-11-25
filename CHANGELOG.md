# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial package scaffolding: `Client`, `Option` constructors, retry loop
  honoring `Retry-After`, typed errors with `apperrors` interop, and per-resource
  service structs (`Standings`, `Scoreboard`, `Conferences`, `Teams`,
  `Programs`, `People`, `Brackets`, `Broadcasts`).
- Standings coverage: `Standings.Get`.
- Scoreboard coverage: `Scoreboard.Day`.
- Conferences coverage: `Conferences.List`, `Conferences.Get`,
  `Conferences.Tournament`.
- Teams coverage: `Teams.List`, `Teams.Get`.
- Programs coverage: `Programs.Get`, `Programs.Roster`.
- People coverage: `People.Journey`, `People.JourneyByID`.
- Brackets coverage: `Brackets.Tournament`, `Brackets.Projected`.
- Broadcasts coverage: `Broadcasts.Day`.
- Error predicates: `IsNotFound`, `IsPaymentRequired`, `IsOutOfScope`,
  `IsRateLimited`.
- `examples/basic` demonstrating the typed-error branch set.
