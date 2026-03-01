package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

type NewsBotData struct {
	BotID       uuid.UUID `json:"botID" dynamodbav:"BotID,binary"`
	URL         string    `json:"url" dynamodbav:"URL,binary"`
	Title       string    `json:"title" dynamodbav:"Title,string"`
	Source      string    `json:"source" dynamodbav:"Source,string"`
	Description string    `json:"description" dynamodbav:"Description,string"`
	Published   time.Time `json:"published" dynamodbav:"Published,number"`
}

func (b *NewsBotData) Key() map[string]any { return map[string]any{"BotID": b.BotID, "URL": b.URL} }

func (b *NewsBotData) Name() string {
	return "ByteLyon_" + ModeTitle() + "_News_Bot_Data"
}

func (b *NewsBotData) Desc() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("BotID"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("URL"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("BotID"),
			AttributeType: types.ScalarAttributeTypeB,
		}, {
			AttributeName: Ptr("URL"),
			AttributeType: types.ScalarAttributeTypeS,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}

func (b *NewsBotData) Validate() error { return nil }
