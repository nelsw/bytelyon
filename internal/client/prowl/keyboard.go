package prowl

import (
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func (c *Client) Type(page playwright.Page, s string) error {
	err := page.Keyboard().Type(s, playwright.KeyboardTypeOptions{
		Delay: Ptr(Between(500.0, 1000.0)),
	})
	log.Err(err).Str("text", s).Msg("Client - Keyboard#Type")
	return err
}

func (c *Client) Press(page playwright.Page, s string) (err error) {
	err = page.Keyboard().Press(s, playwright.KeyboardPressOptions{
		Delay: Ptr(Between(200, 500.0)),
	})
	log.Err(err).Str("key", s).Msg("Client - Keyboard#Press")
	return err
}
