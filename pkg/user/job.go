package user

import (
	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Run(
	pwc *playwright.Playwright,
	uid ulid.ULID,
) {

	var bro playwright.Browser
	var ctx playwright.BrowserContext
	var err error

	for _, t := range bot.Types {
		for _, b := range bot.Find(uid, t) {

			l := log.With().
				Any("type", t).
				Str("target", b.Target).
				Logger()

			l.Trace().Send()

			if !b.IsReady() {
				l.Debug().Msg("bot not ready")
			} else if bro, err = pw.NewBrowser(pwc, b.Headless); err != nil {
				l.Err(err).Msgf("failed to create browser for %s", uid)
			} else if ctx, err = pw.NewBrowserContext(bro, b.Fingerprint); err != nil {
				l.Err(err).Msg("failed to create browser context")
			} else {
				b.Run(bro, ctx, uid)
			}
		}
	}
}
