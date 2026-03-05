package model

import (
	"fmt"
	"regexp"
)

var validationRegex = regexp.MustCompile(`^(search|news|sitemap)$`)
var ErrBotTypeFn = func(a any) error {
	return fmt.Errorf("invalid bot type; must be one of [search, news, or sitemap]; got: [%s]", a)
}

type BotType string

const (
	SearchBotType  BotType = "search"
	SitemapBotType BotType = "sitemap"
	NewsBotType    BotType = "news"
)

func (t BotType) is(bt BotType) bool { return t == bt }
func (t BotType) IsNews() bool       { return t.is(NewsBotType) }
func (t BotType) IsSearch() bool     { return t.is(SearchBotType) }
func (t BotType) IsSitemap() bool    { return t.is(SitemapBotType) }

func (t BotType) Validate() error {
	if validationRegex.MatchString(t.String()) {
		return nil
	}
	return ErrBotTypeFn(t)
}

func (t BotType) String() string {
	return string(t)
}
