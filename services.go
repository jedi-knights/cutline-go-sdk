package cutline

// Service structs. Each is a typed accessor on Client — its methods live in
// the per-resource file (standings.go, scoreboard.go, etc.). Keeping the
// struct declarations together makes the API surface visible in one glance.

// StandingsService accesses RPI standings for a sport+division.
// Obtain it from Client.Standings.
type StandingsService struct{ c *Client }

// ScoreboardService accesses daily scores and schedules.
// Obtain it from Client.Scoreboard.
type ScoreboardService struct{ c *Client }

// ConferencesService accesses the conference registry and tournament brackets.
// Obtain it from Client.Conferences.
type ConferencesService struct{ c *Client }

// TeamsService accesses team detail — roster, schedule, metadata.
// Obtain it from Client.Teams.
type TeamsService struct{ c *Client }

// ProgramsService accesses school-level program detail and roster.
// Obtain it from Client.Programs.
type ProgramsService struct{ c *Client }

// PeopleService accesses player career journey (college → pro affiliations).
// Obtain it from Client.People.
type PeopleService struct{ c *Client }

// BracketsService accesses conference and projected tournament brackets.
// Obtain it from Client.Brackets.
type BracketsService struct{ c *Client }

// BroadcastsService accesses broadcast / streaming metadata for scheduled games.
// Obtain it from Client.Broadcasts.
type BroadcastsService struct{ c *Client }
