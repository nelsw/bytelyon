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
	str, ok := src.(string)
	if !ok {
		return errors.New("invalid src; must be string")
	}
	*t = BotType(str)
	return nil
}

func (t *BotType) Value() (driver.Value, error) {
	s := string(*t)
	if !validationRegex.MatchString(s) {
		return "", errors.New("invalid bot type; must be one of [search, news, or sitemap]; got: [" + s + "]")
	}
	return s, nil
}

func (t BotType) String() string {
	return string(t)
}
