package model

import (
	"fmt"
	"strings"
)

type SitemapBot struct {
	Bot
}

func (b *SitemapBot) Validate() error {
	if ok := strings.HasPrefix(b.Target, "https://"); !ok {
		return fmt.Errorf("bad url, must begin with https://")
	}
	return nil
}
