package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/service/db"
)

var fake *gofakeit.Faker

func init() {
	fake = gofakeit.New(0)
	config.Init()
	logger.Init()
	db.Migrate(
	//&model.BotNews{},
	//&model.BotNewsResult{},
	//&model.BotSearch{},
	//&model.BotSearchResult{},
	//&model.BotSitemap{},
	//&model.BotSitemapResult{},
	//&model.Email{},
	//&model.Password{},
	//&model.Token{},
	//&model.User{},
	)
}

func Test_Init(t *testing.T) {
	fmt.Println("Test_Init", time.Now().UnixMilli())
}
