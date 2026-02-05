package client

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/client/article"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Logger = logger.Make(gin.TestMode)
}

func TestArticleClient(t *testing.T) {

	query := "ev fire blanket"
	after := time.Now().Add(time.Hour * 24 * 365 * -2)

	c := article.NewClient(query, after)
	arr, err := c.Fetch()

	assert.Nil(t, err)
	assert.True(t, len(arr) > 0)
	util.PrettyPrintln(arr)
}
