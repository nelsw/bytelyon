package news

import (
	"net/http"
	"slices"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/page"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return HandleGet(r)
	case http.MethodDelete:
		return HandleDelete(r)
	}
	return r.NI()
}

func HandleDelete(r api.Request) api.Response {
	if r.URL() != "" {
		m, err := Find(r.UserID(), r.Topic())
		if err != nil {
			return r.NC()
		}
		if err = page.Delete(r.URL(), m[r.URL()].ID); err != nil {
			return r.BAD(err)
		}
		delete(m, r.URL())
		Save(r.UserID(), r.Topic(), m)
		return r.NC()
	}

	if err := Delete(r.UserID(), r.Topic()); err != nil {
		return r.BAD(err)
	}

	return r.NC()
}

func HandleGet(r api.Request) api.Response {
	m, err := Find(r.UserID(), r.Topic())
	if err != nil {
		return r.NC()
	}

	if r.URL() != "" {
		var a Article
		if a, err = page.FindObject[Article](r.URL(), m[r.URL()].ID); err != nil {
			return r.NC()
		}
		return r.OK(a)
	}

	var arr []*Headline
	for k, v := range m {
		v.URL = k
		arr = append(arr, v)
	}
	slices.SortFunc(arr, func(a, b *Headline) int {
		return b.ID.Compare(a.ID)
	})
	return r.OK(arr)
}
