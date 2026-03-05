package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/internal/util"
)

type BotNewsResult struct {
	Model
	Target      string    `json:"target" dynamodbav:"Target,string"`
	URL         string    `json:"url" dynamodbav:"ID,string"`
	Title       string    `json:"title" dynamodbav:"Title,string"`
	Source      string    `json:"source" dynamodbav:"Source,string"`
	Description string    `json:"description" dynamodbav:"Description,string"`
	Published   time.Time `json:"published" dynamodbav:"Published,number"`
}

func (b BotNewsResult) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("target"), KeyType: types.KeyTypeHash},
			{AttributeName: Ptr("id"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("target"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {
			AttributeName: Ptr("id"),
			AttributeType: types.ScalarAttributeTypeS,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}
