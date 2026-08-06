package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

type RestController struct {}

func NewRestController() *RestController {
	return &RestController{}
}

func (r *RestController) Index(ctx http.Context) http.Response {
	return ctx.Response().Success().Json(http.Json{
		"Hello": "Goravel",
	})
}
