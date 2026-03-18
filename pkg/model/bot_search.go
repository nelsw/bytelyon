package model

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
)

// Search represents a Search Bot configuration.
type Search struct {
	// Bot is the base bot entity.
	Bot
	// BlackList is a list of domains that should be excluded from the news results.
	BlackList []string
	// Headless is a flag indicating whether the bot should run in headless mode.
	Headless bool
	// State is the browser state of the bot, containing cookies and origins.
	State BroCtxState
}

func (b *Search) Ignore() map[string]bool {
	var m = make(map[string]bool)
	for _, s := range b.BlackList {
		m[s] = true
	}
	return m
}

func (b *Search) StoragePath(n any, ext string) string {
	return fmt.Sprintf("users/%s/bots/search/%s/%s/%d.%s",
		b.UserID, b.Target, b.ID, n, ext)
}

func (b *Search) Put() *dynamodb.PutItemInput {
	item, _ := attributevalue.MarshalMap(&b)
	return &dynamodb.PutItemInput{
		TableName: SearchBotType.TableName(),
		Item:      item,
	}
}

func (b *Search) MarshalDynamoDBAttributeValue() (value types.AttributeValue, err error) {

	var m map[string]types.AttributeValue
	if m, err = attributevalue.MarshalMap(&b.State); err != nil {
		return
	} else if value, err = b.Bot.MarshalDynamoDBAttributeValue(); err != nil {
		return
	}

	if len(m) > 0 {
		value.(*types.AttributeValueMemberM).Value["state"] = &types.AttributeValueMemberM{Value: m}
	}
	if len(b.BlackList) > 0 {
		value.(*types.AttributeValueMemberM).Value["blackList"] = &types.AttributeValueMemberSS{Value: b.BlackList}
	}
	value.(*types.AttributeValueMemberM).Value["headless"] = &types.AttributeValueMemberBOOL{Value: b.Headless}

	return
}

func (b *Search) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	m := v.(*types.AttributeValueMemberM).Value
	if m == nil {
		log.Warn().Any("v", v).Msg("news bot unmarshal value was nil!")
		return nil
	}
	if val, ok := m["state"]; ok {
		if err = attributevalue.UnmarshalMap(val.(*types.AttributeValueMemberM).Value, &b.State); err != nil {
			return
		}
	}
	if val, ok := m["blackList"]; ok {
		b.BlackList = val.(*types.AttributeValueMemberSS).Value
	}
	if val, ok := m["headless"]; ok {
		b.Headless = val.(*types.AttributeValueMemberBOOL).Value
	}
	return b.Bot.UnmarshalDynamoDBAttributeValue(v)
}

func (b *Search) String() string {
	bs := b.Bot.String()
	ns := fmt.Sprintf("\tBlackList: %v\n", b.BlackList)
	ns += fmt.Sprintf("\tHeadless: %v\n", b.Headless)
	ns += fmt.Sprintf("\tState: %s\n", b.State)
	return bs[:len(bs)-1] + ns + "}"
}
