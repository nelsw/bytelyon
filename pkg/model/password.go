package model

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	UserID ulid.ULID
	Hash   []byte
}

// Compare compares a bcrypt hashed password with its possible
// plaintext equivalent. Returns nil on success, or an error on failure.
func (p *Password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(text))
}

// Generate returns the bcrypt hash of the password at the given cost.
func (p *Password) Generate(text string) (err error) {
	p.Hash, err = bcrypt.GenerateFromPassword([]byte(text), bcrypt.MinCost)
	return
}

func (p *Password) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: Ptr("Password"),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("userId"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("userId"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}
func (p *Password) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: Ptr("Password"),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: p.UserID.String()},
		},
	}
}
func (p *Password) Put() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: Ptr("Password"),
		Item: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: p.UserID.String()},
			"hash":   &types.AttributeValueMemberB{Value: p.Hash},
		},
	}
}

func (p *Password) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {
	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("password unmarshal value was nil")
	} else if p.UserID, err = ulid.ParseStrict(m["userId"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse userId: %w", err)
	}
	p.Hash = m["hash"].(*types.AttributeValueMemberB).Value
	return
}
