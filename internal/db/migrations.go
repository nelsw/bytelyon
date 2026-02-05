package db

import (
	"github.com/nelsw/bytelyon/internal/model"
)

var Migrations = []any{
	&model.User{},
	&model.Job{},
}
