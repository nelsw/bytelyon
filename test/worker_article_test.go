package test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/nelsw/bytelyon/internal/worker/article"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Logger = logger.Make(gin.TestMode)
}

func TestArticle(t *testing.T) {

	arr := article.New(&model.Job{
		Type:   model.ArticleType,
		Target: "ev fire blanket",
	}).Work()

	assert.True(t, len(arr) > 0)
	util.PrettyPrintln(arr)
}
