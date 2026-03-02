package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
)

type SitemapBotData struct {
	UserID    uuid.UUID `json:"userID" dynamodbav:"UserID,binary"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"CreatedAt,number"`
	URL       string    `json:"url" dynamodbav:"URL,string"`
	Domain    string    `json:"domain" dynamodbav:"Domain,string"`
	Relative  []string  `json:"relative" dynamodbav:"Relative,stringset"`
	Remote    []string  `json:"remote" dynamodbav:"Remote,stringset"`
}

func (b SitemapBotData) Desc() dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("UserID"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("CreatedAt"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("UserID"),
			AttributeType: types.ScalarAttributeTypeB,
		}, {
			AttributeName: Ptr("CreatedAt"),
			AttributeType: types.ScalarAttributeTypeN,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}
