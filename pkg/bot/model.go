package bot

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/pkg/news"
	"github.com/nelsw/bytelyon/pkg/search"
	"github.com/nelsw/bytelyon/pkg/sitemap"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	typeRegex = regexp.MustCompile(`^(search|news|sitemap)$`)
	Types     = []Type{News, Search, Sitemap}
)

type Type string

const (
	News    = "news"
	Search  = "search"
	Sitemap = "sitemap"
)

type Models []*Model

func (m Models) Len() int      { return len(m) }
func (m Models) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m Models) Less(i, j int) bool {
	if m[i].Type != m[j].Type {
		return strings.Compare(string(m[i].Type), string(m[j].Type)) == -1
	}
	return strings.Compare(m[i].Target, m[j].Target) == -1
}

// Model stores bot configuration and state.
type Model struct {

	// Blacklist is a set of keywords that should be excluded from results.
	Blacklist map[string]bool

	// Fingerprint is the browser state of the bot, containing cookies and origins.
	Fingerprint *playwright.StorageState

	// Frequency is the rate at which to run the bot.
	Frequency time.Duration

	// Headless is a flag indicating whether the bot should run in headless mode.
	Headless bool

	// RanAt is the last time the bot was run.
	RanAt time.Time

	// Target of the bot (e.g., query, domain, etc.).
	Target string

	// Type of bot.
	Type Type
}

func (m *Model) UnmarshalJSON(b []byte) error {

	alias, err := json.Deserialize[struct {
		Blacklist   []string                 `json:"blacklist"`
		Headless    bool                     `json:"headless"`
		Fingerprint *playwright.StorageState `json:"fingerprint,omitempty"`
		Frequency   int64                    `json:"frequency"`
		Target      string                   `json:"target"`
		Type        Type                     `json:"type"`
		RanAt       string                   `json:"ranAt"`
		UserID      ulid.ULID                `json:"userId"`
	}](b)

	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal bot")
		return err
	}

	m.Blacklist = make(map[string]bool)
	for _, k := range alias.Blacklist {
		m.Blacklist[k] = true
	}
	m.Fingerprint = alias.Fingerprint
	m.Frequency = time.Duration(alias.Frequency)
	m.Headless = alias.Headless
	m.RanAt, _ = time.Parse(time.RFC3339, alias.RanAt)
	m.Target = alias.Target
	m.Type = alias.Type
	return nil
}

func (m *Model) MarshalJSON() (b []byte, err error) {
	var blacklist []string
	if len(m.Blacklist) > 0 {
		blacklist = slices.Sorted(maps.Keys(m.Blacklist))
	} else {
		blacklist = make([]string, 0)
	}
	b = json.Of(
		"blacklist", blacklist,
		"fingerprint", m.Fingerprint,
		"frequency", m.Frequency.Nanoseconds(),
		"headless", m.Headless,
		"ranAt", m.RanAt.Format(time.RFC3339),
		"target", m.Target,
		"type", m.Type,
	)
	return
}

// IsReady returns true if the frequency is positive and the next run is in the past.
func (m *Model) IsReady() bool {
	return m.Frequency > 0 && m.RanAt.Add(m.Frequency).Before(time.Now().UTC())
}

func (m *Model) Run(
	bro playwright.Browser,
	ctx playwright.BrowserContext,
	uid ulid.ULID,
) {

	defer func() {
		if err := ctx.Close(); err != nil {
			log.Err(err).Msg("failed to close browser context")
		}
		if err := bro.Close(); err != nil {
			log.Err(err).Msg("failed to close browser")
		}
	}()

	switch m.Type {
	case News:
		news.Work(ctx, uid, m.Target, m.Blacklist, m.RanAt)
	case Search:
		search.Work(ctx, uid, m.Target, m.Blacklist)
	case Sitemap:
		sitemap.Work(ctx, uid, m.Target)
	}

	// update bot worked at to now
	m.RanAt = time.Now().UTC()

	// reset frequency if set to 1ns (once & stop)
	if m.Frequency == 1 {
		m.Frequency = 0
	}

	// update the storage state of the bot
	if state, err := ctx.StorageState(); err != nil || state == nil {
		log.Warn().Err(err).Msg("failed to get browser storage state")
	} else {
		m.Fingerprint = state
	}

	// save bot
	if err := save(uid, m); err != nil {
		log.Warn().Err(err).Msg("failed to Save Search Bot")
	}
}

func (m *Model) Validate() error {
	if !typeRegex.MatchString(string(m.Type)) {
		return fmt.Errorf("invalid bot type; need [search, news, sitemap]; got: [%s]", m.Type)
	}
	return nil
}
