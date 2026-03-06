package model

import (
	"fmt"
)

var ErrBotTypeFn = func(a any) error {
	return fmt.Errorf("invalid bot type; must be one of [search, news, or sitemap]; got: [%s]", a)
}

type Type interface {
	Validate() error
	Table() *string
	String() string
}

func DetermineType(s string) (Type, error) {
	switch s {
	case "search":
		return Search{}, nil
	case "news":
		return News{}, nil
	case "sitemap":
		return Sitemap{}, nil
	default:
		return nil, ErrBotTypeFn(s)
	}
}
