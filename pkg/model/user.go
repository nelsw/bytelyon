package model

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

// User represents a user entity in the system.
type User struct {
	// ID is the unique identifier for the user
	ID ulid.ULID `json:"id"`
}

func (u *User) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: Ptr("ByteLyon_User"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: u.ID.String()},
		},
	}
}

func (u *User) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {
	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("user unmarshal value was nil")
	} else if u.ID, err = ulid.ParseStrict(m["id"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}
	return
}
