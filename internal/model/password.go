package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	ID   uuid.UUID `json:"-" dynamodbav:"ID,binary"`
	Hash []byte    `json:"-" dynamodbav:"Hash,binary"`
}

// Authenticate returns nil if the given plaint text value is equivalent to this Password.Hash, or an error on failure.
func (p *Password) Authenticate(text string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(text))
}

func (p *Password) Desc() *dynamodb.CreateTableInput {
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
		TableName: TableName(p),
	}
}

func NewPassword(userID uuid.UUID, text string) *Password {
	return &Password{
		ID:   userID,
		Hash: Must(bcrypt.GenerateFromPassword([]byte(text), bcrypt.MinCost)),
	}
}

func (p *Password) Update(text string) {
	p.Hash = Must(bcrypt.GenerateFromPassword([]byte(text), bcrypt.MinCost))
}
