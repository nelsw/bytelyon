package test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func Test_DB_Find(t *testing.T) {
	//db.Save(&BotSearch{
	//	Bot: Bot{
	//		Model:     Make(),
	//		Target:    "https://www.associateweb-enabled.info/mission-critical/eyeballs",
	//		Type:      "",
	//		Frequency: 0,
	//		BlackList: nil,
	//	},
	//	Headless: false,
	//	State:    BroCtxState{},
	//})
	arr, scanErr := db.Scan[BotNewsResult](BotNewsResult{})
	util.PrettyPrintln(arr)
	assert.NoError(t, scanErr)
	assert.NotEmpty(t, arr)
	out, findErr := db.Find[BotNews](map[string]any{
		"UserID": arr[0].UserID,
		"Target": arr[0].Target,
	})
	assert.NoError(t, findErr)
	assert.NotNil(t, out)
	util.PrettyPrintln(out)
}

func Test_DB_Query(t *testing.T) {

	var err error

	//arr, scanErr := db.Scan[BotSearch](BotSearch{})
	//assert.NoError(t, scanErr)
	//assert.NotEmpty(t, arr)

	arr, err := db.Query[BotNewsResult](BotNewsResult{}, "how bad is the iran war today")
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)
	util.PrettyPrintln(arr)
}

func Test_DB_Scan(t *testing.T) {
	arr, err := db.Scan[BotSearch](BotSearch{})
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)
	util.PrettyPrintln(arr)
}

func Test_DB_Save(t *testing.T) {
	var userIDs = []uuid.UUID{
		uuid.New(),
		uuid.New(),
	}
	for i := 0; i < 10; i++ {

		bot := Bot{
			Model:     Model{UserID: userIDs[i%2]},
			BlackList: []string{fake.DomainName()},
			Frequency: time.Hour * time.Duration(fake.Uint8()),
			Target:    fake.URL(),
		}
		if i < 3 {
			db.Save(&BotSearch{Bot: bot, Headless: i%2 == 0})
		} else if i < 5 {
			db.Save(&BotSitemap{Bot: bot})
		} else {
			db.Save(&BotNews{Bot: bot})
		}
	}

}

func Test_DB_Wipe(t *testing.T) {
	arr, err := db.Scan[BotSearch](BotSearch{})
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)

	err = db.Wipe(BotSearch{}, Data{
		"UserID": arr[0].UserID,
		"Target": arr[0].Target,
	})
	assert.NoError(t, err)
}
