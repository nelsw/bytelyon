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

	r.Use(gin.Recovery(), gin.Logger(), cors.New(cfg))

	api := r.Group("/api", ValidateAuth)
	{
		api.Group("/user").
			POST("/login", Login).
			POST("/forgot-password", ForgotPassword).
			POST("/signup", Signup).
			POST("/token/:token", Token)
		// todo - delete account
	}
	{
		api.Group("/bots/:type").
			PUT("", SaveBot).
			GET("", GetBots).
			DELETE("/bot/:botID", DeleteBot)
		{
			api.Group("/data/:dataID").
				DELETE("", DeleteBotData).
				GET("", GetBotData)
		}
	}
	return r
}
