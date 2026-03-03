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

func BotEntities() []Entity {
	return []Entity{
		NewsBotType.BotEntity(),
		SearchBotType.BotEntity(),
		SitemapBotType.BotEntity(),
	}
}

func (t BotType) BotEntity(a ...any) Entity {
	switch t {
	case SearchBotType:
		return &BotSearch{Bot: Bot{Model: Make(a...)}}
	case SitemapBotType:
		return &BotSitemap{Bot: Bot{Model: Make(a...)}}
	case NewsBotType:
		return &BotNews{Bot: Bot{Model: Make(a...)}}
	}
	return nil
}

func (t BotType) ResultEntity() Entity {
	switch t {
	case SearchBotType:
		return &BotSearchResult{}
	case SitemapBotType:
		return &BotSitemapResult{}
	case NewsBotType:
		return &BotNewsResult{}
	}
	return nil
}

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

func NewBotType(s string) (BotType, error) {
	if err := BotType(s).Validate(); err != nil {
		return "", err
	}
	return BotType(s), nil
}

func (t BotType) String() string {
	return string(t)
}
