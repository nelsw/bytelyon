package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
)

type Token struct {
	ID     uuid.UUID `json:"id" dynamodbav:"ID,binary"`
	UserID uuid.UUID `json:"userId" dynamodbav:"UserID,string"`
	Type   TokenType `json:"type" dynamodbav:"Type,string"`
	Expiry time.Time `json:"expiry" dynamodbav:"Expiry,number"`
}

func (t Token) Desc() dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("ID"),
			KeyType:       types.KeyTypeHash,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("ID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}

func NewResetPasswordToken(userID uuid.UUID) *Token {
	return &Token{
		ID:     Must(uuid.NewV7()),
		UserID: userID,
		Type:   ResetPasswordTokenType,
		Expiry: time.Now().Add(15 * time.Minute),
	}
}

func NewConfirmEmailToken(userID uuid.UUID) *Token {
	return &Token{
		ID:     Must(uuid.NewV7()),
		UserID: userID,
		Type:   ConfirmEmailTokenType,
		Expiry: time.Now().Add(time.Hour),
	}
}
