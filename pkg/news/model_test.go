package news

import (
	"testing"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
)

func TestModel_Run(t *testing.T) {

	logs.Init("trace")

	cpw := pw.Run()
	bro, _ := pw.NewBrowser(cpw, true)
	ctx, _ := pw.NewBrowserContext(bro, nil)

	defer func() {
		ctx.Close()
		bro.Close()
	}()

	m := New(ulid.Zero, "situation in iran")
	m.Run(ctx, map[string]bool{
		"google": true,
	})

}
