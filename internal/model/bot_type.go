package model

import (
	"database/sql/driver"
	"errors"
	"regexp"
)

var validationRegex = regexp.MustCompile(`^(search|news|sitemap)$`)

type BotType string

const (
	SearchBotType  BotType = "search"
	SitemapBotType BotType = "sitemap"
	NewsBotType    BotType = "news"
)

func (t *BotType) Scan(src any) error {
	if src == nil {
		return errors.New("nil bot type")
	}
	*t = BotType(src.(string))
	if err := t.Validate(); err != nil {
		return err
	}
	return nil
}

func (t BotType) String() string {
	return string(t)
}

func (t *BotType) Validate() error {
	if !validationRegex.MatchString(t.String()) {
		return errors.New("invalid bot type; must be one of [search, news, or sitemap]; got: [" + t.String() + "]")
	}
	return nil
}

func (t *BotType) Value() (driver.Value, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t.String(), nil
}
