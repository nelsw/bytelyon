package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
)

var fake *gofakeit.Faker

func init() {
	config.Init()
	logger.Init()
	fake = gofakeit.New(0)
}
