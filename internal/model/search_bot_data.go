package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

type SearchBotData struct {
	BotID  uuid.UUID        `json:"botID" dynamodbav:"BotID,binary"`
	DataID uuid.UUID        `json:"dataID" dynamodbav:"DataID,binary"`
	Pages  []map[string]any `json:"pages" dynamodbav:"Pages,omitempty"`
}

func (b *SearchBotData) Desc() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName:   TableName(b),
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
