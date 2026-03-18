package model

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/oklog/ulid/v2"
)

type BotSearchResult struct {
	Model
	ID     ulid.ULID  `json:"URL" dynamodbav:"URL,binary"`
	Target string     `json:"target" dynamodbav:"Target,string"`
	Pages  []PageData `json:"pages" dynamodbav:"Pages,omitempty"`
}

func (b BotSearchResult) GetDesc() dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("Target"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: Ptr("URL"),
			KeyType:       types.KeyTypeRange,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("Target"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {
			AttributeName: Ptr("URL"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
	}
}

func (b *BotSearchResult) S3Key(url, ext string) string {
	return fmt.Sprintf("users/%s/bots/search/%d/%s.%s",
		b.UserID,
		b.CreatedAt.UnixMilli(),
		base64.URLEncoding.EncodeToString([]byte(url)),
		ext,
	)
}
