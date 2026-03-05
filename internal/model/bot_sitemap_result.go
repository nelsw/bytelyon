package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
)

type BotSitemapResult struct {
	Model
	ID       ulid.ULID `json:"ID" dynamodbav:"ID,binary"`
	Target   string    `json:"target" dynamodbav:"Target,string"`
	Relative []string  `json:"relative" dynamodbav:"Relative,stringset"`
	Remote   []string  `json:"remote" dynamodbav:"Remote,stringset"`
}

func (b BotSitemapResult) GetDesc() dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("Target"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("ID"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("Target"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {
			AttributeName: Ptr("ID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}
