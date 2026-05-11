package em

import (
	"sort"
	"time"

	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/store"
)

func SavePages(page entity.Pages) {
	for _, p := range page {
		SavePage(p)
	}
}

func SavePage(page *entity.Page) {
	savePageData(page)
	savePageImage(page)
}

func savePageData(page *entity.Page) {
	db, err := store.New[string, entity.Pages]("pages", page.URL, "data")
	if err != nil {
		return
	}
	defer db.Close()

	k := page.ID.Timestamp().Format(time.RFC3339)

	e, ok := db.Get(k)
	if ok {
		e = append(e, page)
	} else {
		e = entity.Pages{page}
	}
	sort.Sort(e)
	db.Set(k, e)
}

func savePageImage(page *entity.Page) {
	db, err := store.New[string, []byte]("pages", page.URL, "screenshots", ".png")
	if err != nil {
		return
	}
	defer db.Close()
	db.Set(page.ID.String(), page.Screenshot)
}

func GetPageData(url, ts string) (*entity.Page, bool) {
	// todo - presigned url
	e, ok := GetPagesData(url)
	if !ok {
		return nil, false
	}
	for _, p := range e {
		if p.ID.String() == ts {
			return p, true
		}
	}
	return nil, false
}

func GetPagesData(url string) (entity.Pages, bool) {
	db, err := store.New[string, *entity.Page]("pages", url, "data")
	if err != nil {
		return nil, false
	}
	defer db.Close()
	return db.Values(), true
}
