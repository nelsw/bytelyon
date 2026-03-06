package model

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type BotAlias struct {
	UserID    string      `json:"userID"`
	Target    Target      `json:"target"`
	Type      Type        `json:"type"`
	Frequency string      `json:"frequency"`
	BlackList []string    `json:"blackList"`
	Headless  bool        `json:"headless"`
	State     BroCtxState `json:"state"`
	UpdatedAt string      `json:"updatedAt"`
	CreatedAt string      `json:"createdAt"`
}

type Bot[T Type] struct {
	BotID     ulid.ULID
	UserID    ulid.ULID
	T         Type
	Target    Target
	Frequency time.Duration
	BlackList []string
	UpdatedAt time.Time
}

func (b Bot[T]) Validate() error {
	// todo: validate frequency
	err := b.T.Validate()
	return err
}

func (b Bot[T]) IsReady() bool {
	return b.Frequency > 0 && b.UpdatedAt.Add(b.Frequency).Before(time.Now().UTC())
}

func (b Bot[T]) Ignore() map[string]bool {
	var m = make(map[string]bool)
	for _, s := range b.BlackList {
		m[s] = true
	}
	return m
}

func (b Bot[T]) Scan() *dynamodb.ScanInput {
	return &dynamodb.ScanInput{
		TableName: b.T.Table(),
	}
}
func (b Bot[T]) Query() *dynamodb.QueryInput {
	keyEx := expression.Key("userID").Equal(expression.Value(b.UserID.Bytes()))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		log.Err(err).Msg("failed to build expression")
		return nil
	}
	return &dynamodb.QueryInput{
		TableName:                 b.T.Table(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
}
func (b Bot[T]) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: b.T.Table(),
		Key: map[string]types.AttributeValue{
			"userID": &types.AttributeValueMemberB{Value: b.UserID.Bytes()},
			"target": &types.AttributeValueMemberS{Value: b.Target.String()},
		},
	}
}
func (b Bot[T]) Put() *dynamodb.PutItemInput {
	if b.Frequency == 1 {
		b.Frequency = 0
	}
	item, err := attributevalue.MarshalMap(&b)
	if err != nil {
		log.Err(err).Msg("failed to marshal bot to dynamodb item!")
	}
	return &dynamodb.PutItemInput{
		TableName: b.T.Table(),
		Item:      item,
	}
}

func (b Bot[T]) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: b.T.Table(),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("userID"), KeyType: types.KeyTypeHash},
			{AttributeName: Ptr("target"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("userID"), AttributeType: types.ScalarAttributeTypeB},
			{AttributeName: Ptr("target"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}

func (b *Bot[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(&BotAlias{
		UserID:    b.UserID.String(),
		Target:    b.Target,
		Type:      b.T,
		Frequency: b.Frequency.String(),
		BlackList: b.BlackList,
		CreatedAt: b.BotID.Timestamp().Format(time.RFC3339Nano),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (b *Bot[T]) UnmarshalJSON(data []byte) (err error) {

	var alias BotAlias
	if err = json.Unmarshal(data, &alias); err != nil {
		return err
	} else if b.Frequency, err = time.ParseDuration(alias.Frequency); err != nil {
		return err
	} else if b.UpdatedAt, err = time.Parse(time.RFC3339Nano, alias.UpdatedAt); err != nil {
		return err
	} else if b.UserID, err = ulid.Parse(alias.UserID); err != nil {
		return err
	}

	b.Target = alias.Target
	b.T = alias.Type
	b.BlackList = alias.BlackList

	return
}

func (b *Bot[T]) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {

	m := map[string]types.AttributeValue{
		"userID":    &types.AttributeValueMemberB{Value: b.UserID.Bytes()},
		"target":    &types.AttributeValueMemberS{Value: b.Target.String()},
		"type":      &types.AttributeValueMemberS{Value: b.T.String()},
		"frequency": &types.AttributeValueMemberS{Value: b.Frequency.String()},
		"blackList": &types.AttributeValueMemberSS{Value: b.BlackList},
		"createdAt": &types.AttributeValueMemberS{Value: b.BotID.Timestamp().Format(time.RFC3339Nano)},
		"updatedAt": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339Nano)},
	}

	return &types.AttributeValueMemberM{Value: m}, nil
}

func (b *Bot[T]) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Msg("bot unmarshal value was nil!")
		return nil
	}

	_ = b.UserID.UnmarshalBinary(m["userID"].(*types.AttributeValueMemberB).Value)
	b.Target, _ = ConstructTarget(m["target"].(*types.AttributeValueMemberS).Value)
	b.T, _ = DetermineType(m["type"].(*types.AttributeValueMemberS).Value)
	b.Frequency, _ = time.ParseDuration(m["frequency"].(*types.AttributeValueMemberS).Value)
	b.BlackList = m["blackList"].(*types.AttributeValueMemberSS).Value
	b.UpdatedAt, _ = time.Parse(time.RFC3339Nano, m["updatedAt"].(*types.AttributeValueMemberS).Value)

	return nil
}
