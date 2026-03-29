package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

// Email represents a single mail address, the User it belongs to, a token for address confirmation.
type Email struct {

	// Address is a unique email address and primary key of the Email table.
	Address string `json:"address"`

	// UserID is the URL of the User this Email belongs to.
	UserID ulid.ULID `json:"userId"`

	// VerifiedAt is the time when the Email was verified.
	VerifiedAt time.Time `json:"verifiedAt"`
}

func (e *Email) TableName() *string {
	return Ptr("Email")
}

func (e *Email) Scan() *dynamodb.ScanInput {
	return &dynamodb.ScanInput{
		TableName: e.TableName(),
	}
}
func (e *Email) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: e.TableName(),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("address"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("address"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}
func (e *Email) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: e.TableName(),
		Key: map[string]types.AttributeValue{
			"address": &types.AttributeValueMemberS{Value: e.Address},
		},
	}
}
func (e *Email) Put() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: e.TableName(),
		Item: map[string]types.AttributeValue{
			"address":    &types.AttributeValueMemberS{Value: e.Address},
			"userId":     &types.AttributeValueMemberS{Value: e.UserID.String()},
			"verifiedAt": &types.AttributeValueMemberS{Value: e.VerifiedAt.Format(time.RFC3339)},
		},
	}
}

func (e *Email) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {
	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("bot unmarshal value was nil")
	} else if e.UserID, err = ulid.ParseStrict(m["userId"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse userId: %w", err)
	} else if e.VerifiedAt, err = time.Parse(time.RFC3339, m["verifiedAt"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse verifiedAt: %w", err)
	}

	e.Address = m["address"].(*types.AttributeValueMemberS).Value

	return
}
