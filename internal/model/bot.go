package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
)

type Bot struct {
	UserID    uuid.UUID     `json:"userID" dynamodbav:"UserID,binary"`
	Target    string        `json:"target" dynamodbav:"Target,string"`
	Type      BotType       `json:"type" dynamodbav:"Type,string"`
	Frequency time.Duration `json:"frequency" dynamodbav:"Frequency,number"`
	BlackList []string      `json:"blackList" dynamodbav:"BlackList,stringset"`
	UpdatedAt time.Time     `json:"updatedAt" dynamodbav:"UpdatedAt,number"`
}

func (b *Bot) IsReady() bool {
	if b.Frequency == 0 {
		return false
	}
	return b.UpdatedAt.Add(b.Frequency).After(time.Now())
}

func (b *Bot) Ignore() map[string]bool {
	m := map[string]bool{}
	for _, s := range b.BlackList {
		m[s] = true
	}
	return m
}

func (b *Bot) desc() dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("UserID"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("Target"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("UserID"),
			AttributeType: types.ScalarAttributeTypeB,
		}, {
			AttributeName: Ptr("Target"),
			AttributeType: types.ScalarAttributeTypeS,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}
