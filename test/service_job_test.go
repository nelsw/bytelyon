package test

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestServiceJob(t *testing.T) {

	svc := service.NewJobService(db.New(gin.TestMode))

	svc.Save(&model.Job{
		Enabled:   true,
		Type:      model.SearchType,
		Frequency: time.Hour * 24,
		Target:    "ev fire blanket",
		BlackList: []string{"firefibers.com"},
	})

	jobs, err := svc.List()

	assert.NoError(t, err)
	assert.Equal(t, len(jobs), 1)
	assert.True(t, jobs[0].Enabled)
	assert.Equal(t, jobs[0].Type, model.ArticleType)
	assert.Equal(t, jobs[0].Frequency, time.Hour*24)
	assert.Equal(t, jobs[0].Target, "ev fire blanket")
	assert.Equal(t, jobs[0].BlackList, []string{"firefibers.com"})
}
