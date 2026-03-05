package model

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var passwordTable = func() *string { return Ptr(os.Getenv("MODE") + "_ByteLyon_Password") }

type Password struct {
	UserID    ulid.ULID
	Hash      []byte
	CreatedAt time.Time
}

// Compare compares a bcrypt hashed password with its possible
// plaintext equivalent. Returns nil on success, or an error on failure.
func (p Password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(text))
}

// Generate returns the bcrypt hash of the password at the given cost.
func (p Password) Generate(text string) (err error) {
	p.Hash, err = bcrypt.GenerateFromPassword([]byte(text), bcrypt.MinCost)
	return
}

func (p Password) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: passwordTable(),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("userID"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("userID"), AttributeType: types.ScalarAttributeTypeB},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}
func (p Password) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: passwordTable(),
		Key: map[string]types.AttributeValue{
			"userID": &types.AttributeValueMemberB{Value: p.UserID.Bytes()},
		},
	}
}
func (p Password) Put() *dynamodb.PutItemInput {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	item, _ := attributevalue.MarshalMap(p)
	return &dynamodb.PutItemInput{
		TableName: passwordTable(),
		Item:      item,
	}
}
func (p *Password) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
		"userID":    &types.AttributeValueMemberB{Value: p.UserID.Bytes()},
		"hash":      &types.AttributeValueMemberB{Value: p.Hash},
		"createdAt": &types.AttributeValueMemberS{Value: p.CreatedAt.Format(time.RFC3339Nano)},
	}}, nil
}
func (p *Password) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Msg("password unmarshal value was nil!")
		return nil
	}

	_ = p.UserID.UnmarshalBinary([]byte(m["userID"].(*types.AttributeValueMemberS).Value))
	p.Hash = m["hash"].(*types.AttributeValueMemberB).Value
	p.CreatedAt, _ = time.Parse(time.RFC3339Nano, m["createdAt"].(*types.AttributeValueMemberS).Value)

	return nil
}
