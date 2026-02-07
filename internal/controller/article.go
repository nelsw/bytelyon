package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/model"
	"gorm.io/gorm"
)

type ArticleController interface {
	Delete(c *gin.Context)
	Find(c *gin.Context)
}

type articleController struct {
	*gorm.DB
}

func NewArticleController(db *gorm.DB) ArticleController {
	return &articleController{db}
}

func (ctl *articleController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctl.Where("id = ?", uint(id)).Delete(&model.Article{})
	c.Status(http.StatusNoContent)
}

func (ctl *articleController) Find(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var a model.Article

	if ctl.Where("id = ?", uint(id)).First(&a); a.ID == 0 {
		c.Status(http.StatusNoContent)
	} else {
		c.JSON(http.StatusOK, a)
	}

}
