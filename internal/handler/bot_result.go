package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
)

func DeleteResult(c *gin.Context) {
	key := Data{
		"Target": c.Param("target"),
		"ID":     uuid.MustParse(c.Param("id")),
	}
	if err := db.Wipe(botType(c).BotEntity(), key); err != nil {
		errRequest(c, err)
	} else {
		c.Status(http.StatusOK)
	}
}

func ListResults(c *gin.Context) {

	var err error
	var arr any

	if botType(c).IsNews() {
		arr, err = db.Query(BotNewsResult{}, c.Param("target"))
	} else if botType(c).IsSearch() {
		arr, err = db.Query(BotSearchResult{}, c.Param("target"))
	} else {
		arr, err = db.Query(BotSitemapResult{}, c.Param("target"))
	}

	if err != nil {
		errRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, arr)
}
