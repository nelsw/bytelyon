package worker

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/nelsw/bytelyon/internal/worker/article"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Logger = logger.Make(gin.TestMode)
}

func TestArticle(t *testing.T) {

	query := "ev fire blanket"
	after := time.Now().Add(time.Hour * 24 * 365 * -2)

	arr, err := article.New(query, after).Work()

	assert.Nil(t, err)
	assert.True(t, len(arr) > 0)
	util.PrettyPrintln(arr)
}
