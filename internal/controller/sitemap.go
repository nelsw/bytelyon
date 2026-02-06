package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/model"
	"gorm.io/gorm"
)

type SitemapController interface {
	Delete(c *gin.Context)
}

type sitemapController struct {
	*gorm.DB
}

func NewSitemapController(db *gorm.DB) SitemapController {
	return &sitemapController{db}
}

func (ctl sitemapController) Delete(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id parameter"})
	} else if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else if err = ctl.Where("id = ?", uint(id)).Delete(&model.Sitemap{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "sitemap deleted successfully"})
	}
}
