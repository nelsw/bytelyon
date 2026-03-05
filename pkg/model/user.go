package model

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

var userTable = func() *string { return Ptr(os.Getenv("MODE") + "_ByteLyon_User") }

var NewUser = func() *User { return &User{ID: ulid.Make()} }

type User struct {
	ID        ulid.ULID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u *User) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: userTable(),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: u.ID.String()},
		},
	}
}

func (u User) Put() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: userTable(),
		Item: map[string]types.AttributeValue{
			"id":        &types.AttributeValueMemberS{Value: u.ID.String()},
			"createdAt": &types.AttributeValueMemberS{Value: u.CreatedAt.Format(time.RFC3339Nano)},
		},
	}
}
func (u *User) Scan() *dynamodb.ScanInput {
	return &dynamodb.ScanInput{
		TableName: userTable(),
	}
}
func (u *User) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: userTable(),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("id"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: Ptr("createdAt"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}
