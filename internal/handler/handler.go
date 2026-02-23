package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
)

var (
	urlValidationRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
)

func Delete[T any](c *gin.Context) {
	if _, err := db.Builder[T]().Where("id = ?", c.MustGet("ID").(uint)).Delete(c); err != nil {
		panic(err)
	}
}

func ListSearches(c *gin.Context) {

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

func ListSitemaps(c *gin.Context) {

	arr, err := db.Builder[model.Sitemap]().
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

func ListNews(c *gin.Context) {

	arr, err := db.Builder[model.News]().
		Where("bot_id = ?", c.MustGet("ID").(uint)).
		Order("published desc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func ListBots(c *gin.Context) {

	arr, err := db.Builder[model.Bot]().
		Order("target asc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func ListBotsByType(c *gin.Context) {
	t := model.BotType(c.Param("type"))
	if err := t.Validate(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	arr, err := db.Builder[model.Bot]().
		Where("type = ?", t).
		Order("target asc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func CreateBot(c *gin.Context) {
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

	if err := db.Builder[model.Bot]().Create(c, &bot); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bot)
}

func UpdateBot(c *gin.Context) {
	var bot model.Bot
	if err := c.Bind(&bot); err != nil {
		return
	}

	_, err := db.Builder[model.Bot]().
		Where("id = ?", bot.ID).
		Updates(c, bot)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bot)
}

func ValidateID(c *gin.Context) {

	if !strings.Contains(c.FullPath(), ":id") {
		c.Next()
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Set("ID", uint(id))
	c.Next()
}

func FindSettings(c *gin.Context) {
	val, err := db.Builder[model.Settings]().First(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, val)
}

func CreateSettings(c *gin.Context) {

	var val model.Settings
	if err := c.Bind(&val); err != nil {
		return
	}

	if err := db.Builder[model.Settings]().Create(c, &val); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, val)
}

func UpdateSettings(c *gin.Context) {
	var val model.Settings
	if err := c.Bind(&val); err != nil {
		return
	}
	_, err := db.Builder[model.Settings]().
		Where("id = ?", val.ID).
		Updates(c, val)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, val)
}
