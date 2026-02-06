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
}

type articleController struct {
	*gorm.DB
}

func NewArticleController(db *gorm.DB) ArticleController {
	return &articleController{db}
}

func (ctl articleController) Delete(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id parameter"})
	} else if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else if err = ctl.Where("id = ?", uint(id)).Delete(&model.Article{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
	}
}
