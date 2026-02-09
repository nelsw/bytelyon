package handler

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/controller"
	"gorm.io/gorm"
)

func New(mode string, db *gorm.DB) http.Handler {

	gin.SetMode(mode)
	gin.ForceConsoleColor()

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), cors.Default())
	r.Static("/static", "./web")
	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
	}
	{
		ctl := controller.NewJobController(db)
		grp := api.Group("/bots")
		grp.GET("", ctl.List)
		grp.PUT("", ctl.Save)
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/type/:type", ctl.ListWhereType)
	}
	{
		ctl := controller.NewSearchController(db)
		grp := api.Group("/search")
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/bot/:id", ctl.Find)
		{
			// todo - pages
		}
	}
	{
		ctl := controller.NewArticleController(db)
		grp := api.Group("/news")
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/bot/:id", ctl.Find)
	}
	{
		ctl := controller.NewSitemapController(db)
		grp := api.Group("/sitemap")
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/bot/:id", ctl.Find)
	}
	{
		// todo - settings
	}
	return r.Handler()
}
