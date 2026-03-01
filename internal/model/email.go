package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

// Email represents a single mail address, the User it belongs to, a token for address confirmation.
type Email struct {

	// ID is a unique email address and primary key of the Email table.
	ID string `json:"ID" dynamodbav:"ID,string"`

	// UserID is a foreign key reference to define which Email belongs-to a User.
	UserID uuid.UUID `json:"-" dynamodbav:"UserID,binary"`
}

func (e *Email) Desc() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("ID"),
			KeyType:       types.KeyTypeHash,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("ID"),
			AttributeType: types.ScalarAttributeTypeS,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		TableName: TableName(e),
	}
}

func NewEmail(userID uuid.UUID, str string) *Email {
	return &Email{
		ID:     str,
		UserID: userID,
	}
}
