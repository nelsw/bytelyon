package test

import (
	"strings"
	"testing"

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
