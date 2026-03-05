package model

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var sitemapTargetErr = fmt.Errorf("bad url, must begin with https://")

type URL string

type BotAlias struct {
	UserID    string      `json:"id"`
	Target    string      `json:"target"`
	Type      BotType     `json:"type"`
	Frequency string      `json:"frequency"`
	BlackList []string    `json:"blackList"`
	Headless  bool        `json:"headless"`
	State     BroCtxState `json:"state"`
	UpdatedAt string      `json:"updatedAt"`
	CreatedAt string      `json:"createdAt"`
}

type Bot struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    ulid.ULID
	Target    string

	Type      BotType
	Frequency time.Duration
	BlackList []string
	Headless  bool
	State     BroCtxState
}

func (b Bot) Validate() error {
	if b.Type == SitemapBotType && !strings.HasPrefix(b.Target, "https://") {
		return sitemapTargetErr
	} else if err := b.Type.Validate(); err != nil {
		return err
	}
	return nil
}

func (b Bot) IsReady() bool {
	return b.Frequency > 0 && b.UpdatedAt.Add(b.Frequency).Before(time.Now().UTC())
}

func (b Bot) Ignore() map[string]bool {
	var m = make(map[string]bool)
	for _, s := range b.BlackList {
		m[s] = true
	}
	return m
}

func (b *Bot) tableName() *string {
	switch b.Type {
	case SearchBotType:
		return Ptr(os.Getenv("MODE") + "_ByteLyon_Bot_Search")
	case SitemapBotType:
		Ptr(os.Getenv("MODE") + "_ByteLyon_Bot_Sitemap")
	case NewsBotType:
		Ptr(os.Getenv("MODE") + "_ByteLyon_Bot_News")
	}
	log.Warn().Msg("bot type not found?!")
	return nil
}

func (b Bot) Query() *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:              b.tableName(),
		KeyConditionExpression: Ptr("userID = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberB{Value: b.UserID.Bytes()},
		},
	}
}

func (b Bot) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: b.tableName(),
		Key: map[string]types.AttributeValue{
			"userID": &types.AttributeValueMemberB{Value: b.UserID.Bytes()},
			"target": &types.AttributeValueMemberS{Value: b.Target},
		},
	}
}

func (b Bot) Put() *dynamodb.PutItemInput {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	item, _ := attributevalue.MarshalMap(b)
	return &dynamodb.PutItemInput{
		TableName: b.tableName(),
		Item:      item,
	}
}

func (b Bot) Create() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: b.tableName(),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: Ptr("userID"), KeyType: types.KeyTypeHash},
			{AttributeName: Ptr("target"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: Ptr("userID"), AttributeType: types.ScalarAttributeTypeB},
			{AttributeName: Ptr("target"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		BillingMode: types.BillingModeProvisioned,
	}
}

func (b *Bot) MarshalJSON() ([]byte, error) {
	var a = BotAlias{
		UserID:    b.UserID.String(),
		Target:    b.Target,
		Type:      b.Type,
		Frequency: b.Frequency.String(),
		BlackList: b.BlackList,
		CreatedAt: b.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if b.Type == SearchBotType {
		a.Headless = b.Headless
		a.State = b.State
	}
	return json.Marshal(&a)
}

func (b *Bot) UnmarshalJSON(data []byte) (err error) {

	var alias BotAlias
	if err = json.Unmarshal(data, &alias); err != nil {
		return err
	} else if b.Frequency, err = time.ParseDuration(alias.Frequency); err != nil {
		return err
	} else if b.UpdatedAt, err = time.Parse(time.RFC3339Nano, alias.UpdatedAt); err != nil {
		return err
	} else if b.CreatedAt, err = time.Parse(time.RFC3339Nano, alias.CreatedAt); err != nil {
		return err
	} else if b.UserID, err = ulid.Parse(alias.UserID); err != nil {
		return err
	}

	b.Target = alias.Target
	b.Type = alias.Type
	b.BlackList = alias.BlackList

	if b.Type == SearchBotType {
		b.Headless = alias.Headless
		b.State = alias.State
	}

	return
}

func (b *Bot) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {

	m := map[string]types.AttributeValue{
		"userID":    &types.AttributeValueMemberB{Value: b.UserID.Bytes()},
		"target":    &types.AttributeValueMemberS{Value: b.Target},
		"type":      &types.AttributeValueMemberS{Value: b.Type.String()},
		"frequency": &types.AttributeValueMemberS{Value: b.Frequency.String()},
		"blackList": &types.AttributeValueMemberSS{Value: b.BlackList},
		"createdAt": &types.AttributeValueMemberS{Value: b.CreatedAt.Format(time.RFC3339Nano)},
		"updatedAt": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339Nano)},
	}
	if b.Type == SearchBotType {
		m["headless"] = &types.AttributeValueMemberBOOL{Value: b.Headless}
		m["state"] = &types.AttributeValueMemberM{Value: b.State.item()}
	}

	return &types.AttributeValueMemberM{Value: m}, nil
}

func (b *Bot) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Msg("bot unmarshal value was nil!")
		return nil
	}

	_ = b.UserID.UnmarshalBinary(m["userID"].(*types.AttributeValueMemberB).Value)
	b.Target = m["target"].(*types.AttributeValueMemberS).Value
	b.Type = BotType(m["type"].(*types.AttributeValueMemberS).Value)
	b.Frequency, _ = time.ParseDuration(m["frequency"].(*types.AttributeValueMemberS).Value)
	b.BlackList = m["blackList"].(*types.AttributeValueMemberSS).Value
	b.CreatedAt, _ = time.Parse(time.RFC3339Nano, m["createdAt"].(*types.AttributeValueMemberS).Value)
	b.UpdatedAt, _ = time.Parse(time.RFC3339Nano, m["updatedAt"].(*types.AttributeValueMemberS).Value)
	if b.Type == SearchBotType {
		b.Headless = m["headless"].(*types.AttributeValueMemberBOOL).Value
		b.State.unmarshal(m["state"].(*types.AttributeValueMemberM))
	}

	return nil
}
