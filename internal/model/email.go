package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
)

// Email represents a single mail address, the User it belongs to, a token for address confirmation.
type Email struct {
	Model

	// Address is a unique email address and primary key of the Email table.
	Address string `json:"address" dynamodbav:"Address"`
}

func (e Email) GetDesc() dynamodb.CreateTableInput {
	d := e.Model.GetDesc()
	d.KeySchema = []types.KeySchemaElement{{
		AttributeName: Ptr("Address"),
		KeyType:       types.KeyTypeHash,
	}}
	d.AttributeDefinitions = []types.AttributeDefinition{{
		AttributeName: Ptr("Address"),
		AttributeType: types.ScalarAttributeTypeS,
	}}
	return d
}

func NewEmail(userID uuid.UUID, str string) *Email {
	return &Email{Model{UserID: userID}, str}
}
