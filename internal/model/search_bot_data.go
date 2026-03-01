package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

type SearchBotData struct {
	BotID  uuid.UUID `json:"botID" dynamodbav:"BotID,binary"`
	DataID uuid.UUID `json:"dataID" dynamodbav:"DataID,binary"`
	// page uri set?
	// serp result data?
}

func (b *SearchBotData) Key() map[string]any {
	return map[string]any{"BotID": b.BotID, "DataID": b.DataID}
}

func (b *SearchBotData) Name() string { return "ByteLyon_" + ModeTitle() + "_Search_Bot_Data" }

func (b *SearchBotData) Desc() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("BotID"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("DataID"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("BotID"),
			AttributeType: types.ScalarAttributeTypeB,
		}, {
			AttributeName: Ptr("DataID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}

func (b *SearchBotData) Validate() error { return nil }
