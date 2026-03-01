package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/client/dynamodb"
	"github.com/nelsw/bytelyon/internal/model"
)

var (
	badBotType = func(s string) map[string]any {
		return map[string]any{
			"error": `invalid bot type; want [search, news, sitemap]; got: [` + s + `]`,
		}
	}
)

func SaveBot(c *gin.Context) {

	ƒ := func(e client.Entity) {
		if err := c.Bind(&e); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		} else if err = e.Validate(); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		} else if err = client.PutItem(c, db, e); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		} else {
			c.JSON(http.StatusCreated, e)
		}
	}

	switch t := model.BotType(c.Param("type")); t {
	case model.SearchBotType:
		ƒ(&model.SearchBot{})
	case model.SitemapBotType:
		ƒ(&model.SitemapBot{})
	case model.NewsBotType:
		ƒ(&model.NewsBot{})
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, badBotType(c.Param("type")))
	}
}

func GetBots(c *gin.Context) {

	var arr any
	var err error

	switch t := model.BotType(c.Param("type")); t {
	case model.SearchBotType:
		arr, err = client.QueryByID[model.SearchBot](c, db, &model.SearchBot{}, c.MustGet("userID").(uuid.UUID))
	case model.SitemapBotType:
		arr, err = client.QueryByID[model.SitemapBot](c, db, &model.SitemapBot{}, c.MustGet("userID").(uuid.UUID))
	case model.NewsBotType:
		arr, err = client.QueryByID[model.NewsBot](c, db, &model.NewsBot{}, c.MustGet("userID").(uuid.UUID))
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

	bot := model.Bot{
		UserID: c.MustGet("userID").(uuid.UUID),
		BotID:  uuid.MustParse(c.Param("botID")),
	}

	switch t := model.BotType(c.Param("type")); t {
	case model.SearchBotType:
		err = client.DeleteItem[model.Bot](c, db, &model.SearchBot{Bot: bot})
	case model.SitemapBotType:
		err = client.DeleteItem[model.Bot](c, db, &model.SitemapBot{Bot: bot})
	case model.NewsBotType:
		err = client.DeleteItem[model.Bot](c, db, &model.NewsBot{Bot: bot})
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
		} else if err = client.DeleteItem(c, db, &model.NewsBotData{BotID: botID, URL: string(url)}); err != nil {
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
			err = client.DeleteItem(c, db, &model.SearchBotData{BotID: botID, DataID: dataID})
		} else {
			err = client.DeleteItem(c, db, &model.SitemapBotData{BotID: botID, DataID: dataID})
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
		arr, err = client.QueryByID[model.SearchBotData](c, db, &model.SearchBotData{}, botID)
	case model.SitemapBotType:
		arr, err = client.QueryByID[model.SitemapBotData](c, db, &model.SitemapBotData{}, botID)
	case model.NewsBotType:
		arr, err = client.QueryByID[model.NewsBotData](c, db, &model.NewsBotData{}, botID)
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, badBotType(c.Param("type")))
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
	} else {
		c.JSON(http.StatusOK, arr)
	}
}
