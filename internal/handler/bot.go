package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
)

var (
	badBotType = func(s string) map[string]any {
		return map[string]any{
			"error": `invalid bot type; want [search, news, sitemap]; got: [` + s + `]`,
		}
	}
)

func SaveBot(c *gin.Context) {

	t := model.BotType(c.Param("type"))

	if t == model.SitemapBotType {
		var b model.SitemapBot
		if err := c.Bind(&b); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		} else if err = b.Validate(); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		} else if err = db.Save(&b); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		} else {
			c.JSON(http.StatusCreated, &b)
		}
		return
	}

	if t == model.SearchBotType || t == model.NewsBotType {
		var a any
		if err := c.Bind(&a); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		} else if err = db.Save(&a); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		} else {
			c.JSON(http.StatusCreated, &a)
		}
		return
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, badBotType(c.Param("type")))
}

func GetBots(c *gin.Context) {

	var arr any
	var err error

	switch t := model.BotType(c.Param("type")); t {
	case model.SearchBotType:
		arr, err = db.Query[model.SearchBot](model.SearchBot{}, "UserID", c.MustGet("userID").(uuid.UUID))
	case model.SitemapBotType:
		arr, err = db.Query[model.SitemapBot](model.SitemapBot{}, "UserID", c.MustGet("userID").(uuid.UUID))
	case model.NewsBotType:
		arr, err = db.Query[model.NewsBot](model.NewsBot{}, "UserID", c.MustGet("userID").(uuid.UUID))
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, badBotType(c.Param("type")))
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
	} else {
		c.JSON(http.StatusOK, arr)
	}
}

func DeleteBot(c *gin.Context) {

	_, err := uuid.Parse(c.Param("botID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
	}

	bot := map[string]any{
		"UserID": c.MustGet("userID").(uuid.UUID),
		"BotID":  uuid.MustParse(c.Param("botID")),
	}

	switch t := model.BotType(c.Param("type")); t {
	case model.SearchBotType:
		err = db.Wipe(model.SearchBot{}, bot)
	case model.SitemapBotType:
		err = db.Wipe(model.SitemapBot{}, bot)
	case model.NewsBotType:
		err = db.Wipe(model.NewsBot{}, bot)
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, badBotType(c.Param("type")))
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
	} else {
		c.Status(http.StatusOK)
	}
}

func DeleteBotData(c *gin.Context) {

	botID, err := uuid.Parse(c.Param("botID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	switch t := model.BotType(c.Param("type")); t {
	case model.NewsBotType:
		var url []byte
		if url, err = base64.URLEncoding.DecodeString(c.Param("dataID")); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		} else if err = db.Wipe(model.NewsBotData{}, map[string]any{"BotID": botID, "URL": string(url)}); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		} else {
			c.Status(http.StatusOK)
		}
	case model.SearchBotType:
		fallthrough
	case model.SitemapBotType:

		var dataID uuid.UUID
		if dataID, err = uuid.Parse(c.Param("dataID")); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if t == model.SearchBotType {
			err = db.Wipe(model.SearchBotData{}, map[string]any{"BotID": botID, "DataID": dataID})
		} else {
			err = db.Wipe(model.SitemapBotData{}, map[string]any{"BotID": botID, "DataID": dataID})
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		} else {
			c.Status(http.StatusOK)
		}
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, badBotType(c.Param("type")))
	}
}

func GetBotData(c *gin.Context) {

	botID, err := uuid.Parse(c.Param("botID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	var arr any
	switch t := model.BotType(c.Param("type")); t {
	case model.SearchBotType:
		arr, err = db.Query[model.SearchBotData](model.SearchBotData{}, "BotID", botID)
	case model.SitemapBotType:
		arr, err = db.Query[model.SitemapBotData](model.SitemapBotData{}, "BotID", botID)
	case model.NewsBotType:
		arr, err = db.Query[model.NewsBotData](model.NewsBotData{}, "BotID", botID)
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, badBotType(c.Param("type")))
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
	} else {
		c.JSON(http.StatusOK, arr)
	}
}
