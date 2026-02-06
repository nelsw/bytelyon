package test

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/client/prowl"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/nelsw/bytelyon/internal/worker/search"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = logger.Make(gin.TestMode)
	prowl.Init(gin.TestMode)
}

func TestSearch(t *testing.T) {

	a := search.New(&model.Job{
		Enabled:   true,
		Type:      model.SearchType,
		Frequency: time.Hour * 24,
		Target:    "ev fire blanket",
		BlackList: []string{"firefibers.com"},
	}).Work()

	util.PrettyPrintln(a)
}
