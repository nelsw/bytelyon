package search

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/oklog/ulid/v2"
)

func TestModel_Run(t *testing.T) {

	logs.Init("trace")

	cpw := pw.Run()
	bro, _ := pw.NewBrowser(cpw, false)
	ctx, _ := pw.NewBrowserContext(bro, nil)

	defer func() {
		if err := ctx.Close(); err != nil {
			t.Error(err)
		}
		if err := bro.Close(); err != nil {
			t.Error(err)
		}
		if err := cpw.Stop(); err != nil {
			t.Error(err)
		}
	}()

	Work(ctx, ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"), "ev fire blankets for sale", map[string]bool{})
}
