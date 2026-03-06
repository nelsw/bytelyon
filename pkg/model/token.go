package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var tokenTable = func() *string { return Ptr("ByteLyon_Token") }

type TokenType string

const ResetPasswordTokenType TokenType = "reset"
const ConfirmEmailTokenType TokenType = "confirm"

var NewToken = func(userID ulid.ULID, tokenType TokenType) *Token {
	return &Token{
		ulid.Make(),
		userID,
		tokenType,
		time.Now().UTC().Add(30 * time.Minute),
	}
}

type Token struct {
	ID     ulid.ULID `json:"id"`
	UserID ulid.ULID `json:"userID"`
	Type   TokenType `json:"type"`
	Expiry time.Time `json:"expiry"`
}

func (t *Token) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: tokenTable(),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("id"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("id"), AttributeType: types.ScalarAttributeTypeB},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}
func (t *Token) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: tokenTable(),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberB{Value: t.ID.Bytes()},
		},
	}
}
func (t *Token) Put() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: tokenTable(),
		Item: map[string]types.AttributeValue{
			"id":     &types.AttributeValueMemberB{Value: t.ID.Bytes()},
			"userID": &types.AttributeValueMemberB{Value: t.UserID.Bytes()},
			"type":   &types.AttributeValueMemberS{Value: string(t.Type)},
			"expiry": &types.AttributeValueMemberS{Value: t.Expiry.Format(time.RFC3339)},
		},
	}
}

func (t *Token) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Msg("email unmarshal value was nil!")
		return nil
	}

	_ = t.ID.UnmarshalBinary(m["id"].(*types.AttributeValueMemberB).Value)
	_ = t.UserID.UnmarshalBinary(m["userID"].(*types.AttributeValueMemberB).Value)
	t.Type = TokenType(m["type"].(*types.AttributeValueMemberS).Value)
	t.Expiry, _ = time.Parse(time.RFC3339Nano, t.Expiry.Format(time.RFC3339))

	return nil
}
