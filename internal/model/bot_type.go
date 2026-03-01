package model

import (
	"fmt"
	"regexp"
)

var validationRegex = regexp.MustCompile(`^(search|news|sitemap)$`)

type BotType string

const (
	SearchBotType  BotType = "search"
	SitemapBotType BotType = "sitemap"
	NewsBotType    BotType = "news"
)

func (t BotType) Validate() error {
	if validationRegex.MatchString(string(t)) {
		return nil
	}
	return fmt.Errorf("invalid bot type; must be one of [search, news, or sitemap]; got: [%s]", t)
}
