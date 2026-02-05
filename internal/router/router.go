package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Router struct {
	*gin.Engine
	db *gorm.DB
}

func New(mode string, db *gorm.DB) *gin.Engine {

	gin.SetMode(mode)
	gin.ForceConsoleColor()

	r := gin.Default()
	//r := gin.New()

	r.LoadHTMLGlob("web/templates/*")
	//r.Static("/assets", "./assets")
	r.GET("/", func(c *gin.Context) {
		var data any
		c.HTML(http.StatusOK, "page.gohtml", data)
	})

	return r
}
