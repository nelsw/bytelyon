package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"gorm.io/gorm"
)

var (
	urlValidationRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
)

func Delete[T any](c *gin.Context) {
	db.MustDelete[T](func(db *gorm.Statement) { db.Where("id = ?", c.MustGet("ID").(uint)) })
}

func FindSearch(c *gin.Context) {

	arr, err := db.Builder[model.Search]().
		Preload("Bot", nil).
		Preload("Pages", nil).
		Where("bot_id = ?", c.MustGet("ID").(uint)).
		Order("created_at desc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func FindSitemap(c *gin.Context) {

	arr, err := db.Builder[model.Search]().
		Preload("Bot", nil).
		Where("bot_id = ?", c.MustGet("ID").(uint)).
		Order("created_at desc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func FindNews(c *gin.Context) {
	c.JSON(http.StatusOK, db.MustFind[model.News](func(db *gorm.DB) *gorm.DB {
		return db.
			Where("bot_id = ?", c.MustGet("ID").(uint)).
			Order("published desc")
	}))
}

func ListBots(c *gin.Context) {
	c.JSON(http.StatusOK, db.MustFind[*model.Bot]())
}

func ListBotsByType(c *gin.Context) {
	t := model.BotType(c.Param("type"))
	if err := t.Validate(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, db.MustFind[*model.Bot](func(db *gorm.DB) *gorm.DB { return db.Where("type = ?", t) }))
}

func SaveBot(c *gin.Context) {
	db.MustSave(c.MustGet("bot").(*model.Bot))
	c.JSON(http.StatusCreated, c.MustGet("bot").(*model.Bot))
}

func ValidateBot(c *gin.Context) {
	var bot model.Bot
	if err := c.Bind(&bot); err != nil {
		return
	}

	if bot.Type == model.SitemapBotType {
		if ok := urlValidationRegex.MatchString(bot.Target); !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad url, must begin with https://"})
			return
		}
	}

	c.Set("bot", &bot)
	c.Next()
}

func ValidateID(c *gin.Context) {

	if !strings.Contains(c.FullPath(), "/id/:id") {
		c.Next()
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Set("ID", id)
	c.Next()
}
