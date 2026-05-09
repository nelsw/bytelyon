package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
)

type Page struct {
	URL            string
	Title          string
	Meta           Data[string]
	Paragraphs     *Set
	ScreenshotKey  string
	ContentKey     string
	CreatedAt      *Time
	ScreenshotData []byte
	ContentData    string
}

func (p *Page) String() string {
	b, _ := json.MarshalIndent(p, "", "\t")
	return string(b)
}

func (p *Page) MarshalJSON() ([]byte, error) {
	return json.Marshal(Data[any]{
		"url":           p.URL,
		"title":         p.Title,
		"meta":          p.Meta,
		"paragraphs":    p.Paragraphs,
		"screenshotKey": p.ScreenshotKey,
		"contentKey":    p.ContentKey,
		"createdAt":     p.CreatedAt,
	})
}

func (p *Page) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: p.Query().TableName,
		Key: map[string]types.AttributeValue{
			"createdAt": p.CreatedAt.ToAttributeValue(),
			"url":       &types.AttributeValueMemberS{Value: p.URL},
		},
	}
}

func (p *Page) Query() *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:                Ptr("ByteLyon_Page"),
		KeyConditionExpression:   Ptr("#0 = :0"),
		ExpressionAttributeNames: map[string]string{"#0": "url"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":0": &types.AttributeValueMemberS{Value: p.URL},
		},
		ScanIndexForward: Ptr(false),
	}
}

func (p *Page) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	m := map[string]types.AttributeValue{
		"url":           &types.AttributeValueMemberS{Value: p.URL},
		"title":         &types.AttributeValueMemberS{Value: p.Title},
		"meta":          p.Meta.ToAttributeValue(),
		"screenshotKey": &types.AttributeValueMemberS{Value: p.ScreenshotKey},
		"contentKey":    &types.AttributeValueMemberS{Value: p.ContentKey},
		"createdAt":     p.CreatedAt.ToAttributeValue(),
	}
	if p.Paragraphs.Len() > 0 {
		m["paragraphs"] = p.Paragraphs.ToAttributeValue()
	}
	return &types.AttributeValueMemberM{Value: m}, nil
}

func (p *Page) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("page unmarshal value was nil")
	} else if p.CreatedAt, err = ParseTime(m["createdAt"]); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}

	p.URL = m["url"].(*types.AttributeValueMemberS).Value
	p.Title = m["title"].(*types.AttributeValueMemberS).Value
	p.Meta = ParseData[string](m["meta"])
	p.Paragraphs = ParseSet(m["paragraphs"])
	if val, ok := m["screenshotKey"]; ok && val != nil {
		p.ScreenshotKey = val.(*types.AttributeValueMemberS).Value
	}
	if val, ok := m["contentKey"]; ok && val != nil {
		p.ContentKey = val.(*types.AttributeValueMemberS).Value
	}

	return
}
