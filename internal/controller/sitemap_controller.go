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
	Find(c *gin.Context)
}

type sitemapController struct {
	*gorm.DB
}

func NewSitemapController(db *gorm.DB) SitemapController {
	return &sitemapController{db}
}

func (ctl sitemapController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctl.Where("id = ?", uint(id)).Delete(&model.Sitemap{})
	c.Status(http.StatusNoContent)
}

func (ctl sitemapController) Find(c *gin.Context) {
	botID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var arr []model.Sitemap

	ctl.Preload("Bot").
		Where("bot_id = ?", uint(botID)).
		Order("created_at desc").
		Find(&arr)

	if len(arr) == 0 {
		c.Status(http.StatusNoContent)
	} else {
		c.JSON(http.StatusOK, arr)
	}
}
