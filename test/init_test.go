package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	. "github.com/nelsw/bytelyon/internal/util"
)

var fake *gofakeit.Faker

func init() {
	fake = gofakeit.New(0)
	config.Init()
	logger.Init()
	db.Migrate(
		Ptr(model.Email{}).Desc(),
		Ptr(model.NewsBot{}).Desc(),
		Ptr(model.NewsBotData{}).Desc(),
		Ptr(model.Password{}).Desc(),
		Ptr(model.SearchBot{}).Desc(),
		Ptr(model.SearchBotData{}).Desc(),
		Ptr(model.SitemapBot{}).Desc(),
		Ptr(model.SitemapBotData{}).Desc(),
		Ptr(model.Token{}).Desc(),
		Ptr(model.User{}).Desc(),
	)
}
