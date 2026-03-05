package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/pkg/db"
	. "github.com/nelsw/bytelyon/pkg/model"
)

var fake *gofakeit.Faker

func init() {
	fake = gofakeit.New(0)
	config.Init()
	logger.Init()
	db.Migrate(
	//Bot{Type: NewsBotType},
	//Bot{Type: SearchBotType},
	//Bot{Type: SitemapBotType},
	//BotNewsResult{},
	//BotSearch{},
	//BotSitemapResult{},
	//Password{},
	//Token{},
	//Email{},
	//User{},
	)
}

func Test_Init(t *testing.T) {
	fmt.Println("Test_Init", time.Now().UnixMilli(), &Email{})

	now := time.Now().Add(time.Duration(rand.Intn(1000000000)) * time.Millisecond)

	path := "/api/results/:type/target/:target"

	methFn := func(method string) string {
		return logger.WhiteBoldIntense + fmt.Sprintf("%6s", method) + logger.Default
	}

	codeFn := func(code int) string {
		if str := fmt.Sprintf(" %d ", code) + logger.Default; code < 300 {
			return logger.WhiteIntense + logger.GreenBackground + str
		} else if code <= 400 {
			return logger.WhiteIntense + logger.YellowBackground + str
		} else {
			return logger.WhiteIntense + logger.RedBackground + str
		}
	}

	fmtFn := func(path, method string, code int) string {
		return fmt.Sprintf("%s %s %s %s %s %s %v\n",
			logger.BlackIntense+time.Now().Format("15:04:05")+logger.Default,
			logger.WhiteIntense+"GIN",
			logger.BlueIntense+".."+path,
			logger.Cyan+">"+logger.Default,
			methFn(method),
			codeFn(code),
			logger.BlackBoldIntense+time.Since(now).String(),
		)
	}

	fmt.Print(fmtFn(path, "GET", 200))
	fmt.Print(fmtFn(path, "POST", 400))
	fmt.Print(fmtFn(path, "DELETE", 500))
}
