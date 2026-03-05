package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
)

type BotNews struct {
	Bot
	Items map[string]*NewsItem
}

type NewsItem struct {
	URL         []byte    `json:"URL"`
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	Description string    `json:"description"`
	Published   time.Time `json:"published"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (b *NewsItem) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"url":         &types.AttributeValueMemberB{Value: b.URL},
			"title":       &types.AttributeValueMemberS{Value: b.Title},
			"source":      &types.AttributeValueMemberS{Value: b.Source},
			"description": &types.AttributeValueMemberS{Value: b.Description},
			"published":   &types.AttributeValueMemberS{Value: b.Published.Format(time.RFC3339)},
			"createdAt":   &types.AttributeValueMemberS{Value: b.CreatedAt.Format(time.RFC3339Nano)},
		},
	}, nil
}

func (b *NewsItem) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {
	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Msg("bot news item unmarshal value was nil!")
		return nil
	}

	b.URL = m["url"].(*types.AttributeValueMemberB).Value
	b.Title = m["title"].(*types.AttributeValueMemberS).Value
	b.Source = m["source"].(*types.AttributeValueMemberS).Value
	b.Description = m["description"].(*types.AttributeValueMemberS).Value
	b.Published, _ = time.Parse(time.RFC3339, m["published"].(*types.AttributeValueMemberS).Value)
	b.CreatedAt, _ = time.Parse(time.RFC3339Nano, m["createdAt"].(*types.AttributeValueMemberS).Value)

	return nil
}
