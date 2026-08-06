package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support"

	"github.com/nelsw/bytelyon/apps/mgr/app/http/controllers"
	"github.com/nelsw/bytelyon/apps/mgr/app/facades"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("welcome.tmpl", map[string]any{
			"version": support.Version,
		})
	})

	facades.Route().Static("public", "./public")

	userController := controllers.NewRestController()
	facades.Route().Get("/users", userController.Index)
}
