package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var emailTable = func() *string { return Ptr("ByteLyon_Email") }

// Email represents a single mail address, the User it belongs to, a token for address confirmation.
type Email struct {

	// Address is a unique email address and primary key of the Email table.
	Address string `json:"address"`

	// UserID is the ID of the User this Email belongs to.
	UserID ulid.ULID `json:"id"`

	// VerifiedAt is the time when the Email was verified.
	VerifiedAt time.Time `json:"verifiedAt"`
}

func (e Email) Scan() *dynamodb.ScanInput {
	return &dynamodb.ScanInput{
		TableName: emailTable(),
	}
}
func (e Email) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: emailTable(),
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
func (e Email) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: emailTable(),
		Key: map[string]types.AttributeValue{
			"address": &types.AttributeValueMemberS{Value: e.Address},
		},
	}
}
func (e Email) Put() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: emailTable(),
		Item: map[string]types.AttributeValue{
			"address":    &types.AttributeValueMemberS{Value: e.Address},
			"userID":     &types.AttributeValueMemberB{Value: e.UserID.Bytes()},
			"verifiedAt": &types.AttributeValueMemberS{Value: e.VerifiedAt.Format(time.RFC3339Nano)},
			"createdAt":  &types.AttributeValueMemberS{Value: e.UserID.Timestamp().Format(time.RFC3339Nano)},
		},
	}
}

func (e *Email) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Msg("email unmarshal value was nil!")
		return nil
	}

	e.Address = m["address"].(*types.AttributeValueMemberS).Value
	_ = e.UserID.UnmarshalBinary(m["userID"].(*types.AttributeValueMemberB).Value)
	e.VerifiedAt, _ = time.Parse(time.RFC3339Nano, e.VerifiedAt.Format(time.RFC3339Nano))
	return nil
}
