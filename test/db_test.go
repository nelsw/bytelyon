package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func init() {
	//config.Init()
	logger.Init()
	//db.Migrate(&model.BotNewsResult{})
}

func Test_All_Bots(t *testing.T) {
	//
	//var users = make([]model.ID, 5)
	//for i := 0; i < 5; i++ {
	//	users[i] = model.NewID()
	//}
	//var bots = make([]model.ID, 15)
	//for i := 0; i < 15; i++ {
	//	bots[i] = model.NewID()
	//}
	//
	//var targets = make([]string, 10)
	//for i := 0; i < 9; i++ {
	//	targets[i] = gofakeit.Fruit()
	//}
	//
	//for i := 0; i < 30; i++ {
	//	db.Put(&model.NewsItem{
	//		Target:      targets[i%10],
	//		ID:          model.NewID(),
	//		BotID:       bots[i%15],
	//		UserID:      users[i%5],
	//		Title:       gofakeit.Sentence(),
	//		Source:      gofakeit.DomainName(),
	//		Description: gofakeit.Paragraph(),
	//		Published:   time.Now().UTC(),
	//	})
	//}
}

func Test_sUser_DB(t *testing.T) {
	exp := model.NewUser()
	b, err := json.MarshalIndent(exp, "", "\t")
	assert.NoError(t, err)
	fmt.Println(string(b))

	var tav map[string]types.AttributeValue
	tav, err = attributevalue.MarshalMap(&exp)

	var act model.User
	err = attributevalue.UnmarshalMap(tav, &act)
	assert.NoError(t, err)
	assert.Equal(t, exp.ID, act.ID)
}

func Test_User_DBJSON(t *testing.T) {

	var exp = model.Bot{
		UserID:    ulid.Make(),
		Target:    "http://api.giphy.com/v1/gifs/search?q=steve+brule&api_key=mOjI01LHhYCkuVj98whYGRJpMdKWCpVt&limit=5",
		Type:      model.SearchBotType,
		Frequency: time.Hour,
		BlackList: []string{"bitcoin"},
		Headless:  true,
		State: model.BroCtxState{
			Cookies: []model.Cookie{},
			Origins: []model.Origin{},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	b, err := json.MarshalIndent(&exp, "", "\t")
	assert.NoError(t, err)
	fmt.Println(string(b))

	var tav map[string]types.AttributeValue
	tav, err = attributevalue.MarshalMap(&exp)
	util.PrettyPrintln(tav)
	var act model.Bot
	err = attributevalue.UnmarshalMap(tav, &act)
	t.Logf("%+v", act)

	assert.NoError(t, err)
	assert.Equal(t, exp.Target, act.Target)
	assert.Equal(t, exp.UserID, act.UserID)
	assert.Equal(t, exp.Type, act.Type)
	assert.Equal(t, exp.Frequency, act.Frequency)
	assert.Equal(t, exp.BlackList, act.BlackList)
	assert.Equal(t, exp.Headless, act.Headless)
	assert.Equal(t, exp.State, act.State)
}

func Test_User_DB(t *testing.T) {

	usr := model.NewUser()
	b, _ := json.Marshal(usr)
	fmt.Println(string(b))
	assert.NoError(t, db.Put(usr))
	assert.NotEmpty(t, usr.ID)

	b, err := json.Marshal(usr)
	assert.NoError(t, err)
	assert.NotEmpty(t, b)
	t.Log(string(b))

	var att model.User
	att, err = db.Get[model.User](&model.User{ID: usr.ID})
	t.Log(att)
	assert.NoError(t, err)
	assert.Equal(t, usr.ID, att.ID)

	out, e := db.Scan[model.User](&model.User{})
	fmt.Println(len(out), e)
	fmt.Println(out[0])
}
