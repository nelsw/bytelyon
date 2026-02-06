package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/controller"
	"gorm.io/gorm"
)

func New(mode string, db *gorm.DB) http.Handler {

	gin.SetMode(mode)
	gin.ForceConsoleColor()

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	api := r.Group("/api")
	{
		ctl := controller.NewJobController(db)
		grp := api.Group("/jobs")
		grp.GET("", ctl.List)
		grp.PUT("", ctl.Save)
		grp.DELETE("/:id", ctl.Delete)
	}
	{
		ctl := controller.NewArticleController(db)
		grp := api.Group("/articles")
		grp.DELETE("/:id", ctl.Delete)
	}
	{
		ctl := controller.NewSitemapController(db)
		grp := api.Group("/sitemaps")
		grp.DELETE("/:id", ctl.Delete)
	}
	{
		ctl := controller.NewSearchController(db)
		grp := api.Group("/searches")
		grp.DELETE("/:id", ctl.Delete)
	}

	return r.Handler()
}
