package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
)

func CreateBot(c *gin.Context) {
	saveBot(c, botType(c).BotEntity(userID(c)))
}

func UpdateBot(c *gin.Context) {
	saveBot(c, botType(c).BotEntity(userID(c), time.Now()))
}

func saveBot(c *gin.Context, e model.Entity) {

	if err := c.Bind(e); err != nil {
		badRequest(c, err)
		return
	}

	if err := db.Save(e); err != nil {
		errRequest(c, err)
	}

	c.JSON(http.StatusCreated, e)
}

func ListBots(c *gin.Context) {

	var err error
	var arr any

	if botType(c).IsNews() {
		arr, err = db.Query(model.BotNews{}, userID(c))
	} else if botType(c).IsSearch() {
		arr, err = db.Query(model.BotSearch{}, userID(c))
	} else {
		arr, err = db.Query(model.BotSitemap{}, userID(c))
	}

	if err != nil {
		errRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, arr)
}

func DeleteBot(c *gin.Context) {

	key := model.Data{
		"UserID": userID(c),
		"Target": c.Param("target"),
	}

	if err := db.Wipe(botType(c).BotEntity(), key); err != nil {
		errRequest(c, err)
		return
	}

	c.Status(http.StatusOK)
}
