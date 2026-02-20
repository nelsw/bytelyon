package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestThis(t *testing.T) {
	title := "1970 Dodge Challenger Classic Cars for Sale - Classics on Autotrader"
	_, source, ok := strings.Cut(title, " - ")
	if !ok {
	}

	assert.Equal(t, "Classics on Autotrader", source)
}

type Foo struct {
	Bar *string
}

func TestIt(t *testing.T) {

	f := Foo{Bar: util.Ptr("baz")}
	util.PrettyPrintln(f)
	f.Bar = nil
	util.PrettyPrintln(f)
	f.Bar = util.Ptr("biz")
	util.PrettyPrintln(f)
	//gdb := db.New(gin.DebugMode)
	//rows, err := gdb.
	//	Table("news").
	//	Select("id").
	//	//Group("date(created_at)").
	//	Rows()
	//assert.NoError(t, err)
	//
	//defer rows.Close()
	//for rows.Next() {
	//	var m map[string]any
	//	rows.Scan(&m)
	//	util.PrettyPrintln(m)
	//}
	//
	//var results []model.News
	//gdb.Model(&model.News{}).
	//	Group("date(created_at)").
	//	Find(&results)
	//util.PrettyPrintln(results)
}

func TestUnwrapHtml(t *testing.T) {

	src := `<a href="https://news.google.com/rss/articles/CBMif0FVX3lxTFByYWpmTnhYRHB0QWljX1R3QUl0QUlUTVFjOFo1MVVKeVhNaExMMVBQMXhlWm5iVzFVNS01WEl1TTlPMHpid2d1c1dERmJFTF9zZ0xpZU85V1RkYnBhZFFYeUd4bkVWSF9QcEZ0NFExa2w3SzVtVWZnbEtmaDFuVVk?oc=5" target="_blank">Bitcoin Price Drops Below $80,000 as New Buyers Rush to Accumulate</a>&nbsp;&nbsp;<font color="#6f6f6f">Yahoo Finance</font>`
	fmt.Println(parseHtml(src))

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(src))
	assert.NoError(t, err)

	selection := doc.Find("a")
	assert.Len(t, selection.Nodes, 1)
	assert.Equal(t, "Bitcoin Price Drops Below $80,000 as New Buyers Rush to Accumulate", selection.Text())

	doc, err = goquery.NewDocumentFromReader(strings.NewReader("Bitcoin Price Drops Below $80,000 as New Buyers Rush to Accumulate"))
	assert.NoError(t, err)
	assert.Len(t, doc.Find("a").Nodes, 0)
}

func parseHtml(src string) string {
	idx := strings.Index(src, `</a>`)
	if idx == -1 {
		return src
	}
	src = src[:idx]
	src = src[strings.LastIndex(src, ">")+1:]
	return src
}
