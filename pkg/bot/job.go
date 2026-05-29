package bot

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/news"
	"github.com/nelsw/bytelyon/pkg/search"
	"github.com/nelsw/bytelyon/pkg/sitemap"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Run(
	bro playwright.Browser,
	ctx playwright.BrowserContext,
	m *Model,
	uid ulid.ULID,
) {

	switch m.Type {
	case "news":
		news.Work(ctx, uid, m.Target, m.Blacklist, m.RanAt)
	case "search":
		search.Work(ctx, uid, m.Target, m.Blacklist)
	case "sitemap":
		sitemap.Work(ctx, uid, m.Target)
	default:
		log.Warn().Msgf("bot type [%s] not supported", m.Type)
		return
	}

	if err := ctx.Close(); err != nil {
		log.Err(err).Msg("failed to close browser context")
	}
	if err := bro.Close(); err != nil {
		log.Err(err).Msg("failed to close browser")
	}

	// update bot worked at to now
	m.RanAt = time.Now().UTC()

	// reset frequency if set to 1ns (once & stop)
	if m.Frequency == 0 {
		m.Frequency = -1
	}

	// update the storage state of the bot
	if state, err := ctx.StorageState(); err == nil {
		m.Fingerprint.SetState(state)
	}

	// save bot
	if err := Save(uid, m); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Bot")
	}
}
