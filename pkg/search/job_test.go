package search

import (
	"testing"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
)

func TestModel_Run(t *testing.T) {

	logs.Init("trace")

	cpw := pw.Run()
	bro, _ := pw.NewBrowser(cpw, false)
	ctx, _ := pw.NewBrowserContext(bro, nil)

	defer func() {
		ctx.Close()
		bro.Close()
	}()

	Work(ctx, ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"), "ev fire blanket for sale", map[string]bool{})
}
