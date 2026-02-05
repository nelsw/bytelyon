package test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
)

func TestIt(t *testing.T) {
	DB := db.New(gin.TestMode)

	user := model.User{
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}

	DB.Save(&user)
}
