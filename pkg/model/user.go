package model

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

// User represents a user entity in the system.
type User struct {
	// ID is the unique identifier for the user
	ID ulid.ULID `json:"id"`
	// Name is the name of the user
	Name string `json:"name"`
}

func (u *User) MarshalZerologObject(evt *zerolog.Event) {
	evt.Stringer("id", u.ID).
		Str("user", u.Name)
}

func (u *User) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: u.Create().TableName,
		Key: map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: u.ID.String()},
			"name": &types.AttributeValueMemberS{Value: u.Name},
		},
	}
}
func (u *User) Put() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: u.Create().TableName,
		Item: map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: u.ID.String()},
			"name": &types.AttributeValueMemberS{Value: u.Name},
		},
	}
}
func (u *User) Scan() *dynamodb.ScanInput {
	return &dynamodb.ScanInput{
		TableName: u.Create().TableName,
	}
}
func (u *User) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: Ptr("User"),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("id"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}

func (u *User) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {
	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("user unmarshal value was nil")
	} else if u.ID, err = ulid.ParseStrict(m["id"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}
	u.Name = m["name"].(*types.AttributeValueMemberS).Value
	return
}

func (u *User) String() string {
	return fmt.Sprintf("User {\n"+
		"\tID: %s\n"+
		"\tName: %s\n"+
		"}",
		u.ID,
		u.Name,
	)
}
