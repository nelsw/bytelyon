package model

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

var (
	tokenErr = errors.New("invalid JWT token (either expired or unprocessable")
)

type User struct {
	ID uuid.UUID `json:"id" dynamodbav:"ID,binary"`
}

func (u *User) Desc() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
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
		TableName: Ptr(u.Name()),
	}
}

func (u *User) Key() map[string]any { return map[string]any{"ID": u.ID} }
func (u *User) Name() string        { return "ByteLyon_" + ModeTitle() + "_User" }
func (u *User) Validate() error     { return nil }
