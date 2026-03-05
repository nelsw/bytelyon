package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
)

type Model struct {
	CreatedAt time.Time `json:"createdAt" dynamodbav:"CreatedAt,number"`
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"UpdatedAt,number"`
	UserID    ulid.ULID `json:"userID" dynamodbav:"UserID,binary"`
}

func (m Model) GetDesc() dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("UserID"),
			KeyType:       types.KeyTypeHash,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("UserID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}

func Make(a ...any) Model {

	var m Model

	for i, v := range a {
		if i == 0 {
			m.UserID = v.(ulid.ULID)
		} else if i == 1 {
			m.UpdatedAt = v.(time.Time)
		} else if i == 2 {
			m.CreatedAt = v.(time.Time)
		}
	}

	if m.UserID == ulid.Zero {
		m.UserID = uuid.Must(uuid.NewV7())
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = m.UpdatedAt
	}

	return m
}
