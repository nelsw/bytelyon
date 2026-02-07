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

	api := r.Group("/api")
	{
		ctl := controller.NewJobController(db)
		grp := api.Group("/jobs")
		grp.GET("", ctl.List)
		grp.PUT("", ctl.Save)
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/type/:type", ctl.ListWhereType)
	}
	{
		ctl := controller.NewArticleController(db)
		grp := api.Group("/articles")
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/job/:job", ctl.Find)
	}
	{
		ctl := controller.NewSitemapController(db)
		grp := api.Group("/sitemaps")
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/job/:job", ctl.Find)
	}
	{
		ctl := controller.NewSearchController(db)
		grp := api.Group("/searches")
		grp.DELETE("/id/:id", ctl.Delete)
		grp.GET("/job/:job", ctl.Find)
		{
			// todo - pages
		}
	}
	{
		// todo - settings
	}
	return r.Handler()
}
