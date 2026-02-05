package test

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestJob(t *testing.T) {

	DB := db.New(gin.TestMode)

	DB.Save(&model.Job{
		Enabled:   true,
		Type:      model.ArticleType,
		Frequency: time.Hour * 24,
		Target:    "ev fire blanket",
		BlackList: []string{"firefibers.com"},
	})

	var jobs []*model.Job

	DB.Find(&jobs)

	assert.Equal(t, len(jobs), 1)
	assert.True(t, jobs[0].Enabled)
	assert.Equal(t, jobs[0].Type, model.ArticleType)
	assert.Equal(t, jobs[0].Frequency, time.Hour*24)
	assert.Equal(t, jobs[0].Target, "ev fire blanket")
	assert.Equal(t, jobs[0].BlackList, []string{"firefibers.com"})
}
