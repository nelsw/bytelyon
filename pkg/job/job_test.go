package job

import (
	"context"
	"testing"

	"github.com/nelsw/bytelyon/pkg/aws"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/logger"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestWork_Search(t *testing.T) {

	t.Setenv("MODE", "release")
	logger.Init()

	bot, err := db.Scan(&model.Bot{Type: model.SearchBotType})
	assert.NoError(t, err)
	assert.NotEmpty(t, bot)

	New(context.Background(), aws.DB(), aws.S3(), bot[0]).Work()
}
