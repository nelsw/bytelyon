// routes/grpc.go
package routes

import (
	proto "github.com/goravel/example-proto"
	"github.com/nelsw/bytelyon/apps/mgr/app/facades"
	"github.com/nelsw/bytelyon/apps/mgr/app/http/controllers"
)

func Grpc() {
  proto.RegisterUserServiceServer(facades.Grpc().Server(), controllers.NewUserController())
}