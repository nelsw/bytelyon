package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

type SitemapBotData struct {
	BotID    uuid.UUID `json:"botID" dynamodbav:"BotID,binary"`
	DataID   uuid.UUID `json:"dataID" dynamodbav:"DataID,binary"`
	URL      string    `json:"url" dynamodbav:"URL,string"`
	Domain   string    `json:"domain" dynamodbav:"Domain,string"`
	Relative []string  `json:"relative" dynamodbav:"Relative,stringset"`
	Remote   []string  `json:"remote" dynamodbav:"Remote,stringset"`
}

func (b *SitemapBotData) Desc() *dynamodb.CreateTableInput {
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
