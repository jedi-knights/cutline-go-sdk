package cutline

// Standings is the response from GET /v1/standings.
type Standings struct {
	Sport     string     `json:"sport"`
	Division  string     `json:"division"`
	Count     int        `json:"count"`
	Standings []Standing `json:"standings"`
}

// Standing is one team's row in the RPI table.
type Standing struct {
	TeamSlug      string  `json:"teamSlug"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	Ties          int     `json:"ties"`
	Games         int     `json:"games"`
	WinPercentage float64 `json:"winPercentage"`
	// OWP is the mean of opponents' win percentage, with games against this
	// team excluded from each opponent's record.
	OWP float64 `json:"owp"`
	// OOWP is the mean of opponents' opponents' win percentage.
	OOWP float64 `json:"oowp"`
	// RPI = 0.25 * WP + 0.50 * OWP + 0.25 * OOWP.
	RPI float64 `json:"rpi"`
}

// Scoreboard is the response from GET /v1/scoreboard/{sport}/{division}/{year}/{month}/{day}.
type Scoreboard struct {
	Sport    string `json:"sport"`
	Division string `json:"division"`
	Date     string `json:"date"`
	Count    int    `json:"count"`
	Games    []Game `json:"games"`
}

// Game is one game (any state).
type Game struct {
	ID         string   `json:"id"`
	Date       string   `json:"date,omitempty"`
	StartTime  string   `json:"startTime,omitempty"`
	State      string   `json:"state"`
	Home       GameSide `json:"home"`
	Away       GameSide `json:"away"`
	Neutral    bool     `json:"neutral,omitempty"`
	Exhibition bool     `json:"exhibition,omitempty"`
	Venue      string   `json:"venue,omitempty"`
}

// GameSide is one side (home/away) of a game.
type GameSide struct {
	TeamSlug   string `json:"teamSlug"`
	Conference string `json:"conference,omitempty"`
	Score      int    `json:"score"`
	Rank       int    `json:"rank"`
	Winner     bool   `json:"winner"`
}

// ConferencesList is the response from GET /v1/conferences.
type ConferencesList struct {
	Count       int          `json:"count"`
	Conferences []Conference `json:"conferences"`
}

// Conference is one conference registry entry.
type Conference struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

// ConferenceTournament is the response from GET /v1/conferences/{slug}/tournament.
type ConferenceTournament struct {
	Sport      string `json:"sport"`
	Division   string `json:"division"`
	Conference string `json:"conference"`
	Count      int    `json:"count"`
	Games      []Game `json:"games"`
}

// TeamsList is the response from GET /v1/teams.
type TeamsList struct {
	Sport      string `json:"sport"`
	Division   string `json:"division"`
	Conference string `json:"conference,omitempty"`
	Count      int    `json:"count"`
	Teams      []Team `json:"teams"`
}

// Team is one team registry entry (NCAA context).
type Team struct {
	Slug       string `json:"slug"`
	ShortName  string `json:"shortName"`
	Conference string `json:"conference"`
	Sport      string `json:"sport"`
	Division   string `json:"division"`
}

// Program is the response from GET /v1/programs/{slug}: one school-level program.
type Program struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Conference  string `json:"conference"`
	State       string `json:"state"`
	RosterURL   string `json:"rosterUrl"`
	ScheduleURL string `json:"scheduleUrl"`
	StatsURL    string `json:"statsUrl"`
}

// Roster is the response from GET /v1/programs/{slug}/roster.
type Roster struct {
	ProgramSlug string   `json:"programSlug"`
	Season      string   `json:"season,omitempty"`
	Count       int      `json:"count"`
	Players     []Player `json:"players"`
}

// Player is one roster row.
type Player struct {
	PersonID     string `json:"personId,omitempty"`
	JerseyNumber string `json:"jerseyNumber"`
	Name         string `json:"name"`
	Position     string `json:"position"`
	ClassYear    string `json:"classYear,omitempty"`
	Height       string `json:"height,omitempty"`
	Hometown     string `json:"hometown,omitempty"`
	// BirthDate is the player's date of birth as "YYYY-MM-DD". Set
	// on pro rosters (ESPN + NWSL SDP); omitted on NCAA rows where
	// only cohort-year estimation is available.
	BirthDate    string `json:"birthDate,omitempty"`
	GamesPlayed  int    `json:"gamesPlayed,omitempty"`
	GamesStarted int    `json:"gamesStarted,omitempty"`
	Minutes      int    `json:"minutes,omitempty"`
	Goals        int    `json:"goals,omitempty"`
}

// Journey is the response from GET /v1/people/{slug}/journey and
// /v1/people/by-id/{id}/journey: one person and their career stints.
type Journey struct {
	Person Person  `json:"person"`
	Stints []Stint `json:"stints"`
}

// Person is one canonical person across sources, clubs, and levels.
type Person struct {
	Slug         string `json:"slug"`
	DisplayName  string `json:"displayName"`
	BirthYear    int    `json:"birthYear,omitempty"`
	PrimarySport string `json:"primarySport,omitempty"`
}

// Stint is one affiliation stint on a career journey.
type Stint struct {
	Organization     string `json:"organization"`
	OrganizationSlug string `json:"organizationSlug"`
	Competition      string `json:"competition"`
	Season           string `json:"season"`
	Jersey           string `json:"jersey,omitempty"`
	Position         string `json:"position,omitempty"`
}

// Tournament is the response from GET /v1/tournament: one NCAA national
// tournament, round-grouped. When Projected is true the rounds are a
// projection (the real bracket has not been published yet).
type Tournament struct {
	Sport      string            `json:"sport"`
	Division   string            `json:"division"`
	FieldSize  int               `json:"fieldSize"`
	Start      string            `json:"start,omitempty"`
	TotalGames int               `json:"totalGames"`
	Rounds     []TournamentRound `json:"rounds"`
	Projected  bool              `json:"projected,omitempty"`
}

// TournamentRound is one round of an NCAA tournament bracket.
type TournamentRound struct {
	Label        string `json:"label"`
	DisplayLabel string `json:"displayLabel,omitempty"`
	DateRange    string `json:"dateRange,omitempty"`
	GamesTarget  int    `json:"gamesTarget"`
	Games        []Game `json:"games"`
}

// ProjectedTournament is the response from GET /v1/tournament/projected:
// a prediction of the NCAA field as of AsOf.
type ProjectedTournament struct {
	Sport      string           `json:"sport"`
	Division   string           `json:"division"`
	FieldSize  int              `json:"fieldSize"`
	AsOf       string           `json:"asOf"`
	TopSixteen []ProjectionSlot `json:"topSixteen"`
	Unseeded   []ProjectionSlot `json:"unseeded"`
	Bubble     []ProjectionSlot `json:"bubble"`
}

// ProjectionSlot is one slot in the projected tournament field.
type ProjectionSlot struct {
	TeamSlug     string  `json:"teamSlug,omitempty"`
	Source       string  `json:"source"`
	Conference   string  `json:"conference,omitempty"`
	Rank         int     `json:"rank,omitempty"`
	RPI          float64 `json:"rpi,omitempty"`
	InContention bool    `json:"inContention,omitempty"`
}

// Broadcasts is the response from GET /v1/broadcasts/{sport}/{division}/{year}/{month}/{day}.
type Broadcasts struct {
	Sport      string      `json:"sport"`
	Division   string      `json:"division"`
	Date       string      `json:"date"`
	Count      int         `json:"count"`
	Broadcasts []Broadcast `json:"broadcasts"`
}

// Broadcast is per-game broadcast metadata.
type Broadcast struct {
	ID      string `json:"id"`
	Channel string `json:"broadcast,omitempty"`
	URL     string `json:"broadcastUrl,omitempty"`
}
