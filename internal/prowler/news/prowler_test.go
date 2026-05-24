package news

import (
	"testing"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
)

func TestProwler(t *testing.T) {

	logs.Init("debug")

	cpw := pw.Run()
	bro, _ := pw.NewBrowser(cpw, true)
	ctx, _ := pw.NewBrowserContext(bro, nil)

	defer func() {
		ctx.Close()
		bro.Close()
	}()

	e := entity.NewNews(ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ"), "ai today")
	New(e, ctx).Prowl()

}
