package router

import (
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.DebugMode)
	gin.ForceConsoleColor()
}

func New() *gin.Engine {

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	return r
}
