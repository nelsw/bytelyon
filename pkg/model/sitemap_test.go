package model

import (
	"fmt"
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func TestSitemap_Find(t *testing.T) {
	logs.Init("debug")
	e := new(Sitemap).Find(ulid.Zero, "firefibers.com")
	s := string(util.JSON(e))
	fmt.Println(s)

}
