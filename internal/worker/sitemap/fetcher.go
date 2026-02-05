package sitemap

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/rs/zerolog/log"
)

type fetcher struct{}

func (f *fetcher) Crawl(s string) []string {

	r, err := fetch.New(s).Reader()
	if err != nil {
		log.Err(err).Send()
		return nil
	}
	defer r.Close()

	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(r); err != nil {
		log.Err(err).Send()
		return nil
	}

	var ss []string
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		ss = append(ss, sel.AttrOr("href", ""))
	})

	return ss
}
