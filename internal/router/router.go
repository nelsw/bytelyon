package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/handler"
)

func New() *gin.Engine {
	gin.SetMode(config.Mode())
	gin.ForceConsoleColor()

	r := gin.New()

	r.Static("/static", "./web")
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AllowCredentials = true
	cfg.AllowHeaders = append(cfg.AllowHeaders, "Authorization")

	r.Use(gin.Recovery(), cors.New(cfg), Logger())

	api := r.Group("/api", ValidateAuth)
	{
		api.Group("/auth").
			POST("/login", LoginUser).
			POST("/reset", ResetPassword).
			POST("/signup", SignupUser).
			POST("/token/:id", ProcessToken)
	}
	{
		api.Group("/bots/:type", ValidateBotType).
			POST("", CreateBot).
			PUT("", UpdateBot).
			GET("", ListBots).
			DELETE("/id/:id", DeleteBot)
	}
	{
		api.Group("results/:type", ValidateBotType).
			GET("/target/:target", ListResults).
			DELETE("/id/:id", DeleteResult)

	}
	return r
}
