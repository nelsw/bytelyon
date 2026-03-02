package test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func Test_DB_Find(t *testing.T) {
	arr, scanErr := db.Scan[model.SearchBot](model.SearchBot{})
	assert.NoError(t, scanErr)
	assert.NotEmpty(t, arr)

	out, findErr := db.Find[model.SearchBot](map[string]any{
		"UserID": arr[0].UserID,
		"Target": arr[0].Target,
	})
	assert.NoError(t, findErr)
	assert.NotNil(t, out)
	util.PrettyPrintln(out)
}

func Test_DB_Query(t *testing.T) {
	arr, scanErr := db.Scan[model.SearchBot](model.SearchBot{})
	assert.NoError(t, scanErr)
	assert.NotEmpty(t, arr)

	k := "UserID"
	v := arr[0].UserID
	arr, err := db.Query[model.SearchBot](model.SearchBot{}, k, v)

	assert.NoError(t, err)
	assert.NotEmpty(t, arr)
	util.PrettyPrintln(arr)
}

func Test_DB_Scan(t *testing.T) {
	arr, err := db.Scan[model.SearchBot](model.SearchBot{}, "Frequency < :freq", map[string]any{
		":freq": 208800000000000,
	})
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

		bot := model.Bot{
			UserID:    userIDs[i%2],
			BlackList: []string{fake.DomainName()},
			Frequency: time.Hour * time.Duration(fake.Uint8()),
			Target:    fake.URL(),
			UpdatedAt: time.Now(),
		}
		if i < 3 {
			db.Save(&model.SearchBot{Bot: bot, Headless: i%2 == 0})
		} else if i < 5 {
			db.Save(&model.SitemapBot{Bot: bot})
		} else {
			db.Save(&model.NewsBot{Bot: bot})
		}
	}
}

func Test_DB_Wipe(t *testing.T) {
	arr, err := db.Scan[model.SearchBot](model.SearchBot{})
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)

	err = db.Wipe(model.SearchBot{}, map[string]any{
		"UserID": arr[0].UserID,
		"Target": arr[0].Target,
	})
	assert.NoError(t, err)
}
