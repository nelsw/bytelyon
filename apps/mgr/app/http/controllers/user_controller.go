// package controllers

// import (
// 	"github.com/goravel/framework/contracts/http"
// )

// type UserController struct {}

// func NewUserController() *UserController {
// 	return &UserController{}
// }

// func (r *UserController) Index(ctx http.Context) http.Response {
// 	return ctx.Response().Success().Json(http.Json{
// 		"Hello": "Goravel",
// 	})
// }

package controllers

import (
  "context"
  "net/http"

  proto "github.com/goravel/example-proto"
)

type UserController struct {}

func NewUserController() *UserController {
  return &UserController{}
}

func (r *UserController) GetUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
  return &proto.UserResponse{
    Code: http.StatusOK,
    Data: &proto.User{
      Id:    1,
      Name:  "Goravel",
      Token: req.GetToken(),
    },
  }, nil
}