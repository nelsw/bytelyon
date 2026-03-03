package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/internal/util"
)

type Bot struct {
	Model
	Target    string        `json:"target" dynamodbav:"Target,string"`
	Type      BotType       `json:"type" dynamodbav:"Type,string"`
	Frequency time.Duration `json:"frequency" dynamodbav:"Frequency,number"`
	BlackList []string      `json:"blackList" dynamodbav:"BlackList,stringset"`
}

func (b Bot) IsReady() bool {
	if b.Frequency == 0 {
		return false
	}
	return b.UpdatedAt.Add(b.Frequency).Before(time.Now())
}

func (b Bot) Ignore() map[string]bool {
	m := map[string]bool{}
	for _, s := range b.BlackList {
		m[s] = true
	}
	return m
}

func (b Bot) GetDesc() dynamodb.CreateTableInput {
	d := b.Model.GetDesc()
	d.KeySchema = append(d.KeySchema, types.KeySchemaElement{
		AttributeName: Ptr("Target"),
		KeyType:       types.KeyTypeRange,
	})
	d.AttributeDefinitions = append(d.AttributeDefinitions, types.AttributeDefinition{
		AttributeName: Ptr("Target"),
		AttributeType: types.ScalarAttributeTypeS,
	})
	return d
}
