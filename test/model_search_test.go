package test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestModelSearch(t *testing.T) {
	DB := db.New(gin.DebugMode)
	var arr []model.Search
	err := DB.Model(&model.Search{}).Preload("Pages").Find(&arr).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)
	util.PrettyPrintln(arr)

	a := arr[0]
	err = DB.Delete(&a).Error
}
