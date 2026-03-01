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
	BotID     uuid.UUID     `json:"botID" dynamodbav:"BotID,binary"`
	Type      BotType       `json:"type" dynamodbav:"Type,string"`
	Target    string        `json:"target" dynamodbav:"Target,string"`
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

func (b *Bot) Key() map[string]any { return map[string]any{"UserID": b.UserID, "BotID": b.BotID} }

func (b *Bot) Validate() error { return nil }

func (b *Bot) desc() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("UserID"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("BotID"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("UserID"),
			AttributeType: types.ScalarAttributeTypeB,
		}, {
			AttributeName: Ptr("BotID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}
