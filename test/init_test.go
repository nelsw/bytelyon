package test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
)

var fake *gofakeit.Faker

func init() {
	fake = gofakeit.New(0)
	config.Init()
	logger.Init()
	db.Migrate(
		&model.Email{},
		&model.NewsBot{},
		&model.NewsBotData{},
		&model.Password{},
		&model.SearchBot{},
		&model.SearchBotData{},
		&model.SitemapBot{},
		&model.SitemapBotData{},
		&model.Token{},
		&model.User{},
	)
}

func Test_Init(t *testing.T) {}
