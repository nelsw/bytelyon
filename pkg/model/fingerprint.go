package model

import (
	"github.com/nelsw/bytelyon/pkg/util/ptr"
	"github.com/playwright-community/playwright-go"
)

type Fingerprint struct {
	// Cookies to set for context
	Cookies []playwright.Cookie `json:"cookies"`
	// Origins to set for context
	Origins []playwright.Origin `json:"origin"`
}

func (f *Fingerprint) GetState() (s *playwright.OptionalStorageState) {
	s = &playwright.OptionalStorageState{Origins: f.Origins}
	for _, c := range f.Cookies {
		s.Cookies = append(s.Cookies, playwright.OptionalCookie{
			Name:         c.Name,
			Value:        c.Value,
			URL:          nil,
			Domain:       ptr.OrNil(c.Domain),
			Path:         ptr.OrNil(c.Path),
			Expires:      ptr.OrNil(c.Expires),
			HttpOnly:     ptr.OrNil(c.HttpOnly),
			Secure:       ptr.OrNil(c.Secure),
			SameSite:     c.SameSite,
			PartitionKey: c.PartitionKey,
		})
	}
	return
}
