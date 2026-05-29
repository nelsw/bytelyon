package bot

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

var (
	typeErr = func(t string) error {
		return fmt.Errorf("invalid bot type; need any [search, news, sitemap]; got: [%s]", t)
	}
)

type Models []*Model

func (m Models) Len() int      { return len(m) }
func (m Models) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m Models) Less(i, j int) bool {
	if m[i].Type != m[j].Type {
		return strings.Compare(m[i].Type, m[j].Type) == -1
	}
	return strings.Compare(m[i].Target, m[j].Target) == -1
}

// Model stores bot configuration and state.
type Model struct {

	// Blacklist is a list of keywords that should be excluded from results.
	Blacklist model.Set[string]

	// Fingerprint is the browser state of the bot, containing cookies and origins.
	Fingerprint model.Fingerprint

	// Frequency is the rate at which to run the bot.
	Frequency time.Duration

	// Headless is a flag indicating whether the bot should run in headless mode.
	Headless bool

	// RanAt is the last time the bot was run.
	RanAt time.Time

	// Target of the bot (e.g., query, domain, etc.).
	Target string

	// Type of bot.
	Type string

	UserID ulid.ULID
}

func (m *Model) UnmarshalJSON(b []byte) (err error) {

	var alias struct {
		Blacklist []string  `json:"blacklist"`
		Headless  bool      `json:"headless"`
		Frequency int64     `json:"frequency"`
		Target    string    `json:"target"`
		Type      string    `json:"type"`
		RanAt     string    `json:"ranAt"`
		UserID    ulid.ULID `json:"userId"`
	}

	if err = json.Unmarshal(b, &alias); err != nil {
		return
	}

	m.Blacklist = model.MakeSet(alias.Blacklist...)
	m.Headless = alias.Headless
	m.Frequency = time.Duration(alias.Frequency) * time.Hour
	m.Target = alias.Target
	m.Type = alias.Type
	m.RanAt, _ = time.Parse(time.RFC3339, alias.RanAt)
	m.UserID = alias.UserID
	return
}

func (m *Model) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"blacklist": m.Blacklist.Slice(),
		"headless":  m.Headless,
		"frequency": m.Frequency.Hours(),
		"target":    m.Target,
		"type":      m.Type,
		"ranAt":     m.RanAt.Format(time.RFC3339),
	})
}

func (m *Model) GetState() *playwright.OptionalStorageState {
	var cookies []playwright.OptionalCookie
	if m.Fingerprint.Cookies == nil {
		m.Fingerprint.Cookies = []playwright.Cookie{}
	}
	for _, c := range m.Fingerprint.Cookies {
		cookies = append(cookies, playwright.OptionalCookie{
			Name:         c.Name,
			Value:        c.Value,
			URL:          nil,
			Domain:       util.PtrOrNil(c.Domain),
			Path:         util.PtrOrNil(c.Path),
			Expires:      util.PtrOrNil(c.Expires),
			HttpOnly:     util.Ptr(c.HttpOnly),
			Secure:       util.Ptr(c.Secure),
			SameSite:     c.SameSite,
			PartitionKey: c.PartitionKey,
		})
	}

	if m.Fingerprint.Origins == nil {
		m.Fingerprint.Origins = []playwright.Origin{}
	}

	return &playwright.OptionalStorageState{
		Origins: m.Fingerprint.Origins,
		Cookies: cookies,
	}
}

// IsReady returns true if the frequency is positive and the next run is in the past.
func (m *Model) IsReady() bool {
	return m.Frequency > -1 && m.RanAt.Add(m.Frequency).Before(time.Now().UTC())
}

func (m *Model) SetState(s *playwright.StorageState) {
	m.Fingerprint.Cookies = s.Cookies
	m.Fingerprint.Origins = s.Origins
}

func (m *Model) Validate() error {
	if regexp.MustCompile(`^(search|news|sitemap)$`).MatchString(m.Type) {
		return nil
	}
	return typeErr(m.Type)
}
