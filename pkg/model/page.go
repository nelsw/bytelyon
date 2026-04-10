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
	ScreenshotURL  string
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
		"screenshotUrl": p.ScreenshotURL,
		"contentKey":    p.ContentKey,
		"createdAt":     p.CreatedAt,
	})
}

func (p *Page) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: Ptr("Page"),
		Key: map[string]types.AttributeValue{
			"createdAt": p.CreatedAt.ToAttributeValue(),
			"url":       &types.AttributeValueMemberS{Value: p.URL},
		},
	}
}

func (p *Page) Query() *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:                p.Get().TableName,
		KeyConditionExpression:   Ptr("#0 = :0"),
		ExpressionAttributeNames: map[string]string{"#0": "url"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":0": &types.AttributeValueMemberS{Value: p.URL},
		},
		ScanIndexForward: Ptr(false),
	}
}

func (p *Page) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
		"url":           &types.AttributeValueMemberS{Value: p.URL},
		"title":         &types.AttributeValueMemberS{Value: p.Title},
		"meta":          p.Meta.ToAttributeValue(),
		"paragraphs":    p.Paragraphs.ToAttributeValue(),
		"screenshotUrl": &types.AttributeValueMemberS{Value: p.ScreenshotURL},
		"contentKey":    &types.AttributeValueMemberS{Value: p.ContentKey},
		"createdAt":     p.CreatedAt.ToAttributeValue(),
	}}, nil
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
	p.ScreenshotURL = m["screenshotUrl"].(*types.AttributeValueMemberS).Value
	p.ContentKey = m["contentKey"].(*types.AttributeValueMemberS).Value

	return
}
