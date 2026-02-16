package router

import (
	"net/http"

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
	r.Use(gin.Recovery(), gin.Logger(), cors.Default())

	api := r.Group("/api", ValidateID)
	{
		api.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	}
	{
		api.Group("/bots").
			GET("", ListBots).
			PUT("", ValidateBot, SaveBot).
			DELETE("/id/:id", Delete[model.Bot]).
			GET("/type/:type", ListBotsByType)
	}
	{
		api.Group("/searches").
			DELETE("/id/:id", Delete[model.Search]).
			GET("/id/:id", FindSearch)
	}
	{
		api.Group("/news").
			DELETE("/id/:id", Delete[model.News]).
			GET("/id/:id", FindNews)
	}
	{
		api.Group("/sitemaps").
			DELETE("/id/:id", Delete[model.Sitemap]).
			GET("/id/:id", FindSitemap)
	}
	{
		// todo - settings
	}

	return r
}
