package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

// BotResult represents a Bot Result entity.
type BotResult struct {

	// UserID is the user associated with this bot result.
	UserID ulid.ULID

	// BotID is the hash (partition) key for this table.
	BotID ulid.ULID

	// ID is the range key for this table.
	ID ulid.ULID

	// Target is the target of the bot.
	Target string

	// Type is the type of the bot result.
	Type BotType

	// Data is the result of the bot.
	Data any
}

func (b *BotResult) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: b.TableName(),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("botId"), KeyType: types.KeyTypeHash},
			{AttributeName: Ptr("id"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("botId"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: Ptr("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}

func (b *BotResult) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: b.Type.TableName(),
		Key: map[string]types.AttributeValue{
			"botId": &types.AttributeValueMemberS{Value: b.BotID.String()},
			"id":    &types.AttributeValueMemberS{Value: b.ID.String()},
		},
	}
}

func (b *BotResult) Query() *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:                b.TableName(),
		KeyConditionExpression:   Ptr("#0 = :0"),
		ExpressionAttributeNames: map[string]string{"#0": "botId"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":0": &types.AttributeValueMemberS{Value: b.BotID.String()},
		},
	}
}

func (b *BotResult) TableName() *string {
	return b.Type.TableName("Result")
}

func (b *BotResult) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	m, err := attributevalue.MarshalMap(&b.Data)
	if err != nil {
		return nil, err
	}

	m["userId"] = &types.AttributeValueMemberS{Value: b.UserID.String()}
	m["botId"] = &types.AttributeValueMemberS{Value: b.BotID.String()}
	m["id"] = &types.AttributeValueMemberS{Value: b.ID.String()}
	m["type"] = &types.AttributeValueMemberS{Value: b.Type.String()}
	m["target"] = &types.AttributeValueMemberS{Value: b.Target}

	return &types.AttributeValueMemberM{Value: m}, nil
}

func (b *BotResult) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("bot result unmarshal value was nil")
	}

	if err = attributevalue.UnmarshalMap(m["data"].(*types.AttributeValueMemberM).Value, &b.Data); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	if b.UserID, err = ulid.Parse(m["userId"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}

	if b.BotID, err = ulid.Parse(m["botId"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}

	if b.ID, err = ulid.Parse(m["id"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}

	b.Type = BotType(m["type"].(*types.AttributeValueMemberS).Value)
	b.Target = m["target"].(*types.AttributeValueMemberS).Value

	return
}

func (b *BotResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"userId": b.UserID.String(),
		"botId":  b.BotID.String(),
		"id":     b.ID.String(),
		"type":   b.Type.String(),
		"target": b.Target,
		"data":   b.Data,
	})
}

func (b *BotResult) UnmarshalJSON(data []byte) (err error) {

	var m map[string]any
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	} else if b.UserID, err = ulid.ParseStrict(m["userId"].(string)); err != nil {
		return fmt.Errorf("failed to parse UserID: %w", err)
	} else if b.BotID, err = ulid.ParseStrict(m["botId"].(string)); err != nil {
		return fmt.Errorf("failed to parse BotID: %w", err)
	} else if b.ID, err = ulid.ParseStrict(m["id"].(string)); err != nil {
		return fmt.Errorf("failed to parse BotID: %w", err)
	}

	b.Type = BotType(m["type"].(string))
	b.Target = m["target"].(string)
	b.Data = m["data"]

	return
}

func (b *BotResult) String() string {
	byt, _ := json.MarshalIndent(b, "", "\t")
	return string(byt)
}
