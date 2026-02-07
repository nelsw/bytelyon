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
		return
	}

	ctl.DB.Select("Pages").Where("id = ?", uint(id)).Delete(&model.Search{})
	c.Status(http.StatusNoContent)
}

func (ctl *searchController) Find(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var a model.Search
	if ctl.DB.Preload("Pages").First(&a, id); a.ID == 0 {
		c.Status(http.StatusNoContent)
	} else {
		c.JSON(http.StatusOK, &a)
	}
}
