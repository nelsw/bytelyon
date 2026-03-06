package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type News struct {
	ID     ulid.ULID
	BotID  ulid.ULID
	Target string
	Items  map[URL]Item
	Rules  map[string]bool
}

func IncludeOldNews(b *News, old News) {
	for k, v := range old.Items {
		if _, ok := b.Items[k]; !ok {
			b.Items[k] = v
		}
	}
}

func (b News) Put() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: b.Table(),
		//Item:      b.Items,
	}
}

type Item struct {
	URL         URL       `json:"URL"`
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	Description string    `json:"description"`
	Published   time.Time `json:"published"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (b News) String() string  { return "search" }
func (b News) Table() *string  { return util.Ptr("News_Bot") }
func (b News) Validate() error { return nil }

func (b *Item) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"url":         &types.AttributeValueMemberS{Value: b.URL.String()},
			"title":       &types.AttributeValueMemberS{Value: b.Title},
			"source":      &types.AttributeValueMemberS{Value: b.Source},
			"description": &types.AttributeValueMemberS{Value: b.Description},
			"published":   &types.AttributeValueMemberS{Value: b.Published.Format(time.RFC3339)},
			"createdAt":   &types.AttributeValueMemberS{Value: b.CreatedAt.Format(time.RFC3339Nano)},
		},
	}, nil
}

func (b *Item) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) error {
	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Msg("bot news item unmarshal value was nil!")
		return nil
	}

	b.URL = URL(m["url"].(*types.AttributeValueMemberB).Value)
	b.Title = m["title"].(*types.AttributeValueMemberS).Value
	b.Source = m["source"].(*types.AttributeValueMemberS).Value
	b.Description = m["description"].(*types.AttributeValueMemberS).Value
	b.Published, _ = time.Parse(time.RFC3339, m["published"].(*types.AttributeValueMemberS).Value)
	b.CreatedAt, _ = time.Parse(time.RFC3339Nano, m["createdAt"].(*types.AttributeValueMemberS).Value)

	return nil
}
