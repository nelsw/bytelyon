package model

import (
	"fmt"
	"strings"
)

type BotSitemap struct {
	Bot
}

func (b BotSitemap) Validate() error {
	if ok := strings.HasPrefix(b.Target, "https://"); !ok {
		return fmt.Errorf("bad url, must begin with https://")
	}
	return nil
}
