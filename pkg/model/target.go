package model

import (
	"fmt"
	"strings"
)

type Target interface {
	Validate() error
	String() string
}

type Query string

func (q Query) String() string  { return string(q) }
func (q Query) Validate() error { return nil }

type URL string

func (u URL) String() string { return string(u) }
func (u URL) Validate() error {
	if u == "" {
		return fmt.Errorf("url cannot be empty")
	} else if !strings.HasPrefix(u.String(), "https://") {
		return fmt.Errorf("url must begin with https://")
	}
	return nil
}

func ConstructTarget(target string) (Target, error) {
	switch {
	case strings.HasPrefix(target, "https://"):
		return URL(target), nil
	default:
		return Query(target), nil
	}
}
