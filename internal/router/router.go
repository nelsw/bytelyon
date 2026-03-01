package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/handler"
	"github.com/nelsw/bytelyon/internal/model"
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
	}
	{
		api.Group("/bots").
			GET("", ListBots).
			POST("", CreateBot).
			PUT("", UpdateBot).
			DELETE("/id/:id", ValidateID, Delete[model.Bot]).
			GET("/type/:type", ListBotsByType)
	}
	{
		api.Group("/search", ValidateID).
			DELETE("/id/:id", Delete[model.Search]).
			GET("/bot/:id", ListSearches)
	}
	{
		api.Group("/sitemap", ValidateID).
			DELETE("/id/:id", Delete[model.Sitemap]).
			GET("/bot/:id", ListSitemaps)
	}
	{
		api.Group("/news", ValidateID).
			DELETE("/id/:id", Delete[model.News]).
			GET("/bot/:id", ListNews)
	}
	return r
}
