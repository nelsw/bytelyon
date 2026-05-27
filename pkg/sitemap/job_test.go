package sitemap

import (
	"testing"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
)

func TestWork(t *testing.T) {

	logs.Init("debug")

	cpw := pw.Run()
	bro, _ := pw.NewBrowser(cpw, true)
	ctx, _ := pw.NewBrowserContext(bro, nil)

	defer func() {
		ctx.Close()
		bro.Close()
	}()

	Work(ctx, ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"), "firefibers.com")
}
