package test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	dbClient "github.com/nelsw/bytelyon/internal/client/dynamodb"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func Test_Client_Dynamo_Item_Functions(t *testing.T) {

	var err error
	ctx := context.Background()
	c := util.Must(dbClient.New())

	exp := &model.User{util.Must(uuid.NewV7())}

	err = dbClient.PutItem(ctx, c, exp)
	assert.NoError(t, err)

	var act = model.User{exp.ID}
	_, err = dbClient.GetItem[model.User](ctx, c, exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.ID, act.ID)

	err = dbClient.DeleteItem(ctx, c, exp)
	assert.NoError(t, err)

	act, err = dbClient.GetItem[model.User](ctx, c, exp)
	assert.ErrorAs(t, err, &dbClient.NotFoundEx)
}

func Test_Client_Dynamo_Query(t *testing.T) {

	//var err error
	//ctx := context.Background()
	//dbc := util.Must(dbClient.New())
	//name := "ByteLyon_Test_Bot_News"
	//
	//err = dbClient.DeleteTable(ctx, dbc, name)
	//assert.NoError(t, err)
	//
	//err = dbClient.CreateTable(ctx, dbc, &dynamodb.CreateTableInput{
	//	TableName:   &name,
	//	BillingMode: types.BillingModeProvisioned,
	//	ProvisionedThroughput: &types.ProvisionedThroughput{
	//		ReadCapacityUnits:  util.Ptr(int64(10)),
	//		WriteCapacityUnits: util.Ptr(int64(10)),
	//	},
	//	KeySchema: []types.KeySchemaElement{{
	//		AttributeName: util.Ptr("UserID"),
	//		KeyType:       types.KeyTypeHash,
	//	}, {
	//		AttributeName: util.Ptr("ID"),
	//		KeyType:       types.KeyTypeRange,
	//	}},
	//	AttributeDefinitions: []types.AttributeDefinition{{
	//		AttributeName: util.Ptr("UserID"),
	//		AttributeType: types.ScalarAttributeTypeB,
	//	}, {
	//		AttributeName: util.Ptr("ID"),
	//		AttributeType: types.ScalarAttributeTypeB,
	//	}},
	//})
	//assert.NoError(t, err)
	//
	//type Bot struct {
	//	ID        uuid.UUID     `json:"id" dynamodbav:"ID,binary"`
	//	UserID    uuid.UUID     `json:"userId" dynamodbav:"UserID,binary"`
	//	Frequency time.Duration `json:"frequency" dynamodbav:"Frequency,number"`
	//	BlackList []string      `json:"blackList" dynamodbav:"BlackList,stringset"`
	//	Headless  bool          `json:"headless" dynamodbav:"Headless,boolean"`
	//}
	//
	//ids := []uuid.UUID{
	//	util.Must(uuid.NewV7()),
	//	util.Must(uuid.NewV7()),
	//}
	//
	//for j := 0; j < 2; j++ {
	//	for i := 0; i < 5; i++ {
	//		err = dbClient.PutItem(ctx, dbc, Bot{
	//			UserID:    ids[j],
	//			ID:        util.Must(uuid.NewV7()),
	//			Frequency: time.Hour * time.Duration(fake.Uint8()),
	//			BlackList: []string{
	//				fake.DomainName(),
	//			},
	//			Headless: fake.Bool(),
	//		})
	//		assert.NoError(t, err)
	//	}
	//}
	//
	//var arr []Bot
	//arr, err = dbClient.QueryByID[Bot](ctx, dbc, name, "UserID", ids[0])
	//assert.NoError(t, err)
	//assert.Len(t, arr, 5)
	//util.PrettyPrintln(arr)
}

func Test_Client_Dynamo_Email(t *testing.T) {
	ctx := context.Background()
	c := util.Must(dbClient.New())

	//assert.NoError(t, dbClient.CreateTable(ctx, c, exp))
	//assert.NoError(t, dbClient.PutItem(ctx, c, exp))
	//assert.NoError(t, dbClient.DeleteTable(ctx, c, exp))

	email, err := dbClient.GetItem[model.Email](ctx, c, &model.Email{ID: "kowalski7012@gmail.com"})
	assert.NoError(t, err)
	assert.Equal(t, "kowalski7012@gmail.com", email.ID)

	var user model.User
	user, err = dbClient.GetItem[model.User](ctx, c, &model.User{ID: email.UserID})
	assert.NoError(t, err)
	assert.Equal(t, email.UserID, user.ID)
}

func Test_Client_Dynamo_Token(t *testing.T) {
	ctx := context.Background()
	c := util.Must(dbClient.New())

	//exp := &model.Token{
	//	ID:     util.Must(uuid.NewV7()),
	//	UserID: util.Must(uuid.NewV7()),
	//	Type:   model.ConfirmEmailTokenType,
	//	Expiry: time.Now().Add(time.Hour),
	//}
	//assert.NoError(t, dbClient.DeleteTable(ctx, c, exp))
	//assert.NoError(t, dbClient.CreateTable(ctx, c, exp))
	//assert.NoError(t, dbClient.PutItem(ctx, c, exp))

	act, err := dbClient.GetItem[model.Token](ctx, c, &model.Token{})
	assert.NoError(t, err)
	assert.Equal(t, act.ID, uuid.Nil)
	//assert.Equal(t, exp.ID, act.ID)
	//assert.Equal(t, exp.Type, act.Type)
	//assert.Equal(t, exp.Expiry.Truncate(60*time.Second), act.Expiry.Truncate(60*time.Second))
	//assert.Equal(t, exp.UserID, act.UserID)
}
