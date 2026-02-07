package test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/nelsw/bytelyon/internal/worker/sitemap"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = logger.Make(gin.TestMode)
}

func TestSitemap(t *testing.T) {
	m := sitemap.New(&model.Job{
		Type:   model.SitemapType,
		Target: "https://www.ubicquia.com",
	}).Work()
	util.PrettyPrintln(m)
}
