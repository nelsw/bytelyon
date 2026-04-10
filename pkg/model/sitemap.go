package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/util"
)

type Sitemap struct {
	Domain    string
	URLs      *Set
	createdAt *Time
	updatedAt *Time
}

func (s *Sitemap) String() string {
	b, _ := json.MarshalIndent(s, "", "\t")
	return string(b)
}

func (s *Sitemap) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: util.Ptr("Sitemap"),
		Key: map[string]types.AttributeValue{
			"domain": &types.AttributeValueMemberS{Value: s.Domain},
		},
	}
}

func (s *Sitemap) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	s.updatedAt = Now()
	return &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
		"domain":    &types.AttributeValueMemberS{Value: s.Domain},
		"urls":      s.URLs.ToAttributeValue(),
		"createdAt": s.createdAt.ToAttributeValue(),
		"updatedAt": s.updatedAt.ToAttributeValue(),
	}}, nil
}

func (s *Sitemap) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {
	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("sitemap unmarshal value was nil")
	} else if s.createdAt, err = ParseTime(m["createdAt"]); err != nil {
		return fmt.Errorf("failed to parse createdAt: %w", err)
	} else if s.updatedAt, err = ParseTime(m["updatedAt"]); err != nil {
		return fmt.Errorf("failed to parse updatedAt: %w", err)
	}
	s.Domain = m["domain"].(*types.AttributeValueMemberS).Value
	s.URLs = ParseSet(m["urls"])
	return
}

func NewSitemap(domain string) *Sitemap {
	return &Sitemap{
		Domain:    domain,
		URLs:      NewSet(),
		createdAt: Now(),
	}
}
