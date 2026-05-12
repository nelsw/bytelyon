package model

import (
	"fmt"
	"regexp"

	"github.com/nelsw/bytelyon/pkg/util"
)

var validationRegex = regexp.MustCompile(`^(search|news|sitemap)$`)
var ErrBotTypeFn = func(a any) error {
	return fmt.Errorf("invalid bot type; need any [search, news, sitemap]; got: [%s]", a)
}

type BotType string

const (
	SearchBotType  BotType = "search"
	SitemapBotType BotType = "sitemap"
	NewsBotType    BotType = "news"
)

func (t BotType) Validate() error {
	if validationRegex.MatchString(t.String()) {
		return nil
	}
	return ErrBotTypeFn(t)
}

func (t BotType) String() string {
	return string(t)
}

func (t BotType) TableName(args ...string) *string {
	s := "ByteLyon_Bot_" + util.Capitalize(t.String())
	if len(args) > 0 {
		s += "_" + util.Capitalize(args[0])
	}
	return &s
}

func (t BotType) Plural() string {
	if t == SearchBotType {
		return "searches"
	} else if t == NewsBotType {
		return "news"
	} else if t == SitemapBotType {
		return "sitemaps"
	}
	return "Unknown"
}

func BotTypes() []BotType {
	return []BotType{SearchBotType, NewsBotType, SitemapBotType}
}
