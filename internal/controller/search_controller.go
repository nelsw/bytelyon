package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/model"
	"gorm.io/gorm"
)

type SearchController interface {
	Delete(c *gin.Context)
	Find(c *gin.Context)
}

type searchController struct {
	*gorm.DB
}

func NewSearchController(db *gorm.DB) SearchController {
	return &searchController{db}
}

func (ctl *searchController) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctl.DB.Select("Pages").Where("id = ?", uint(id)).Delete(&model.Search{})
	c.Status(http.StatusNoContent)
}

func (ctl *searchController) Find(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var arr []model.Search
	ctl.Preload("Bot").
		Preload("Pages").
		Where("bot_id = ?", id).
		Order("created_at desc").
		Find(&arr)

	if len(arr) == 0 {
		c.Status(http.StatusNoContent)
	} else {
		c.JSON(http.StatusOK, arr)
	}
}
