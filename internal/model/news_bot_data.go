package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
)

type NewsBotData struct {
	UserID      uuid.UUID `json:"userID" dynamodbav:"UserID,binary"`
	URL         string    `json:"url" dynamodbav:"URL,binary"`
	Title       string    `json:"title" dynamodbav:"Title,string"`
	Source      string    `json:"source" dynamodbav:"Source,string"`
	Description string    `json:"description" dynamodbav:"Description,string"`
	Published   time.Time `json:"published" dynamodbav:"Published,number"`
}

func (b NewsBotData) Desc() dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("UserID"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("URL"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("UserID"),
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
