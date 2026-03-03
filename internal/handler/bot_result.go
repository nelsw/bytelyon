package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
)

func DeleteResult(c *gin.Context) {

	var err error

	var key Data

	if botType(c).IsNews() {
		var b []byte
		b, err = base64.RawURLEncoding.DecodeString(c.Param("id"))
		key = Data{"Target": c.Param("target"), "ID": string(b)}
	} else if botType(c).IsSitemap() {
		var b []byte
		b, err = base64.RawURLEncoding.DecodeString(c.Param("target"))
		var id uuid.UUID
		id, err = uuid.Parse(c.Param("id"))
		key = Data{"Target": string(b), "ID": id}
	} else {
		var id uuid.UUID
		id, err = uuid.Parse(c.Param("id"))
		key = Data{"Target": c.Param("target"), "ID": id}
	}

	if err != nil {
		badRequest(c, err)
	} else if err = db.Wipe(botType(c).ResultEntity(), key); err != nil {
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
		var b []byte
		if b, err = base64.RawURLEncoding.DecodeString(c.Param("target")); err != nil {
			errRequest(c, err)
			return
		}
		arr, err = db.Query(BotSitemapResult{}, string(b))
	}

	if err != nil {
		errRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, arr)
}
