package worker

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/nelsw/bytelyon/internal/worker/sitemap"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = logger.Make(gin.TestMode)
}

func TestSitemap(t *testing.T) {
	url := "https://www.ubicquia.com"
	m := sitemap.New(url).Work()
	util.PrettyPrintln(m)
}
