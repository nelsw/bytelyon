package sitemap

import (
	"testing"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
)

func TestNewProwler(t *testing.T) {

	logs.Init("debug")

	cpw := pw.Run()
	bro, _ := pw.NewBrowser(cpw, true)
	ctx, _ := pw.NewBrowserContext(bro, nil)

	defer func() {
		ctx.Close()
		bro.Close()
	}()

	New("firefibers.com", 1, ctx).Prowl(model.NewULID())
}
