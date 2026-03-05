package model

import (
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var NewID = func() ulid.ULID {
	id, err := ulid.New(
		ulid.Timestamp(time.Now()),
		rand.New(rand.NewSource(time.Now().UnixNano())),
	)
	if err != nil {
		id = ulid.Make()
	}
	return id
}

type ID ulid.ULID

func (id ID) IsNil() bool          { return ulid.ULID(id).IsZero() }
func (id ID) String() string       { return ulid.ULID(id).String() }
func (id ID) CreatedAt() time.Time { return ulid.ULID(id).Timestamp().UTC() }

func (id *ID) MarshalJSON() ([]byte, error) {
	return ulid.ULID(*id).MarshalBinary()
}

func (id *ID) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
		"id":        &types.AttributeValueMemberS{Value: id.String()},
		"createdAt": &types.AttributeValueMemberS{Value: id.CreatedAt().Format(time.RFC3339Nano)},
	}}, nil
}

func (id *ID) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	m := v.(*types.AttributeValueMemberM).Value
	s := m["id"].(*types.AttributeValueMemberS).Value

	var tmp ulid.ULID
	if tmp, err = ulid.ParseStrict(s); err != nil {
		log.Err(err).Msg("failed to parse ulid!")
		return
	}

	*id = ID(tmp)
	return
}
