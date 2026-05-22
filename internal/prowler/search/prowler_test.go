package search

import (
	"testing"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
)

func TestProwler_Prowl(t *testing.T) {

	logs.Init("debug")

	cpw := pw.Run()
	bro, _ := pw.NewBrowser(cpw, false)
	ctx, _ := pw.NewBrowserContext(bro, nil)

	defer func() {
		ctx.Close()
		bro.Close()
	}()

	New(ulid.Zero, "ev fire blanket for sale", nil, ctx).Prowl()
}
