package model

import "github.com/oklog/ulid/v2"

type Search struct {
	ID    ulid.ULID
	Pages []*Page
}

func NewSearch() *Search {
	return &Search{
		ID: NewULID(),
	}
}

func (s *Search) AddPage(p *Page) {
	s.Pages = append(s.Pages, p)
}
