package test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/client/dynamodb"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func Test_Client_Dynamo_Table_Functions(t *testing.T) {

	var err error
	var ok bool
	var names []string

	ctx := context.Background()
	dbc := client.New()
	name := "ByteLyon_Test_Table"

	ok, err = client.TableExists(ctx, dbc, name)
	assert.NoError(t, err)
	assert.False(t, ok)

	err = client.CreateTable(ctx, dbc, &dynamodb.CreateTableInput{
		TableName:   &name,
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  util.Ptr(int64(10)),
			WriteCapacityUnits: util.Ptr(int64(10)),
		},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: util.Ptr("ID"),
			KeyType:       types.KeyTypeHash,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: util.Ptr("ID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
	})
	assert.NoError(t, err)

	ok, err = client.TableExists(ctx, dbc, name)
	assert.NoError(t, err)
	assert.True(t, ok)

	names, err = client.ListTables(ctx, dbc)
	assert.NoError(t, err)
	assert.NotEmpty(t, names)

	err = client.DeleteTable(ctx, dbc, name)
	assert.NoError(t, err)

	ok, err = client.TableExists(ctx, dbc, name)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_Client_Dynamo_Item_Functions(t *testing.T) {

	var err error
	ctx := context.Background()
	c := client.New()
	name := "ByteLyon_Test_User"

	type User struct {
		ID uuid.UUID `json:"id" dynamodbav:"ID,binary"`
	}

	var exp = User{util.Must(uuid.NewV7())}

	err = client.PutItem(ctx, c, name, exp)
	assert.NoError(t, err)

	var act = User{exp.ID}
	act, err = client.GetItem[User](ctx, c, name, exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.ID, act.ID)

	err = client.DeleteItem(ctx, c, name, exp)
	assert.NoError(t, err)

	act, err = client.GetItem[User](ctx, c, name, exp)
	assert.ErrorAs(t, err, &client.NotFoundEx)
}

func Test_Client_Dynamo_Query(t *testing.T) {

	var err error
	ctx := context.Background()
	dbc := client.New()
	name := "ByteLyon_Test_Bot_News"

	err = client.DeleteTable(ctx, dbc, name)
	assert.NoError(t, err)

	err = client.CreateTable(ctx, dbc, &dynamodb.CreateTableInput{
		TableName:   &name,
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  util.Ptr(int64(10)),
			WriteCapacityUnits: util.Ptr(int64(10)),
		},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: util.Ptr("UserID"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: util.Ptr("ID"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: util.Ptr("UserID"),
			AttributeType: types.ScalarAttributeTypeB,
		}, {
			AttributeName: util.Ptr("ID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
	})
	assert.NoError(t, err)

	type Bot struct {
		ID        uuid.UUID     `json:"id" dynamodbav:"ID,binary"`
		UserID    uuid.UUID     `json:"userId" dynamodbav:"UserID,binary"`
		Frequency time.Duration `json:"frequency" dynamodbav:"Frequency,number"`
		BlackList []string      `json:"blackList" dynamodbav:"BlackList,stringset"`
		Headless  bool          `json:"headless" dynamodbav:"Headless,boolean"`
	}

	ids := []uuid.UUID{
		util.Must(uuid.NewV7()),
		util.Must(uuid.NewV7()),
	}

	for j := 0; j < 2; j++ {
		for i := 0; i < 5; i++ {
			err = client.PutItem(ctx, dbc, name, Bot{
				UserID:    ids[j],
				ID:        util.Must(uuid.NewV7()),
				Frequency: time.Hour * time.Duration(fake.Uint8()),
				BlackList: []string{
					fake.DomainName(),
				},
				Headless: fake.Bool(),
			})
			assert.NoError(t, err)
		}
	}

	var arr []Bot
	arr, err = client.QueryByID[Bot](ctx, dbc, name, "UserID", ids[0])
	assert.NoError(t, err)
	assert.Len(t, arr, 5)
	util.PrettyPrintln(arr)
}
