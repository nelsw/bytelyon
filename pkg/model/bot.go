package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Bot represents a bot entity with associated configuration and state.
type Bot struct {

	// UserID is the partition (hash) key.
	// The partition key is part of the table's primary key.
	// It is a hash value used to retrieve items from your table
	// and allocate data across hosts for scalability and availability.
	UserID ulid.ULID

	// ID is the unique identifier of the bot.
	ID ulid.ULID

	// Target is the sort (range) key.
	// You can use a sort key as the second part of a table's primary key.
	// The sort key allows you to sort or search among all items sharing the same partition key.
	Target string

	// Type is the type of the bot.
	Type BotType

	// Frequency is the time interval between bot runs.
	Frequency time.Duration

	// WorkedAt is the time when the bot last ran.
	WorkedAt time.Time

	// BlackList is a list of keywords that should be excluded from results.
	BlackList []string

	// Headless is a flag indicating whether the bot should run in headless mode.
	Headless bool

	// Fingerprint is the browser state of the bot, containing cookies and origins.
	Fingerprint *Fingerprint
}

func (b *Bot) Validate() error {
	if b.Frequency < 0 {
		return errors.New("frequency must be greater than 0")
	} else if err := b.Type.Validate(); err != nil {
		return fmt.Errorf("invalid bot type: %w", err)
	}
	return nil
}

func (b *Bot) StoragePath(n any, ext string) string {
	return fmt.Sprintf("users/%s/bots/%s/%s/%s/%d.%s",
		b.UserID,
		b.Type,
		b.Target,
		b.ID,
		n,
		ext,
	)
}

// IsReady returns true if the bot is ready to run.
func (b *Bot) IsReady() bool {

	// 0ns is what the web app sends to pause runs.
	if b.Frequency == 0 {
		return false
	}

	// 1ns is what the web app sends to run once ASAP.
	if b.Frequency == 1 {
		return true
	}

	// add the frequency to the time this bot was last worked
	next := b.WorkedAt.Add(b.Frequency)

	// if the next run is in the past, it's ready to run
	return next.Before(time.Now().UTC())
}

func (b *Bot) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: b.Type.TableName(),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: b.UserID.String()},
			"target": &types.AttributeValueMemberS{Value: b.Target},
		},
	}
}

func (b *Bot) Query() *dynamodb.QueryInput {
	keyEx := expression.Key("userId").Equal(expression.Value(b.UserID.String()))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		log.Err(err).Msg("failed to build expression")
		return nil
	}
	return &dynamodb.QueryInput{
		TableName:                 b.Type.TableName(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          Ptr(false),
	}
}

// MarshalDynamoDBAttributeValue returns a DynamoDB AttributeValue for the bot.
func (b *Bot) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {

	value := map[string]types.AttributeValue{
		"id":        &types.AttributeValueMemberS{Value: b.ID.String()},
		"userId":    &types.AttributeValueMemberS{Value: b.UserID.String()},
		"target":    &types.AttributeValueMemberS{Value: b.Target},
		"type":      &types.AttributeValueMemberS{Value: b.Type.String()},
		"frequency": &types.AttributeValueMemberS{Value: b.Frequency.String()},
		"workedAt":  &types.AttributeValueMemberS{Value: b.WorkedAt.Format(time.RFC3339)},
	}

	if len(b.BlackList) > 0 {
		value["blackList"] = &types.AttributeValueMemberSS{Value: b.BlackList}
	}

	if b.Type == SearchBotType {
		value["headless"] = &types.AttributeValueMemberBOOL{Value: b.Headless}
		if b.Fingerprint != nil {
			NewFingerprint()
		}
		m, err := attributevalue.MarshalMap(b.Fingerprint)
		if err != nil {
			log.Err(err).Msg("failed to marshal bot fingerprint")
			return nil, err
		}
		value["fingerprint"] = &types.AttributeValueMemberM{Value: m}
	}

	return &types.AttributeValueMemberM{Value: value}, nil
}

// UnmarshalDynamoDBAttributeValue populates the bot from a DynamoDB AttributeValue.
func (b *Bot) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {

	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("bot unmarshal value was nil")
	}

	if b.ID, err = ulid.Parse(m["id"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}

	if b.UserID, err = ulid.ParseStrict(m["userId"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse ulid: %w", err)
	}

	b.Target = m["target"].(*types.AttributeValueMemberS).Value
	b.Type = BotType(m["type"].(*types.AttributeValueMemberS).Value)

	if b.Frequency, err = time.ParseDuration(m["frequency"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse frequency: %w", err)
	}

	if b.WorkedAt, err = time.Parse(time.RFC3339, m["workedAt"].(*types.AttributeValueMemberS).Value); err != nil {
		return fmt.Errorf("failed to parse workedAt: %w", err)
	}

	if val, ok := m["blackList"]; ok {
		b.BlackList = val.(*types.AttributeValueMemberSS).Value
	}
	if val, ok := m["headless"]; ok {
		b.Headless = val.(*types.AttributeValueMemberBOOL).Value
	}
	if val, ok := m["fingerprint"]; ok && val != nil && val.(*types.AttributeValueMemberM).Value != nil {
		if err = attributevalue.UnmarshalMap(val.(*types.AttributeValueMemberM).Value, &b.Fingerprint); err != nil {
			log.Err(fmt.Errorf("failed to unmarshal state of bot fingerprint: %w", err))
		}
	}

	return
}

func (b *Bot) MarshalJSON() ([]byte, error) {

	m := map[string]any{
		"userId":    b.UserID.String(),
		"id":        b.ID.String(),
		"target":    b.Target,
		"type":      b.Type.String(),
		"frequency": b.Frequency.Nanoseconds(),
	}

	if !b.WorkedAt.IsZero() {
		m["workedAt"] = b.WorkedAt.Format(time.RFC3339)
	}

	if len(b.BlackList) > 0 {
		m["blackList"] = b.BlackList
	}

	m["headless"] = b.Headless

	if b.Type == SearchBotType {
		m["fingerprint"] = b.Fingerprint
	}

	return json.Marshal(m)
}

func (b *Bot) UnmarshalJSON(data []byte) (err error) {

	var m map[string]any
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}

	if s, ok := m["userId"]; ok && s != nil && s != "" {
		if b.UserID, err = ulid.ParseStrict(s.(string)); err != nil {
			return fmt.Errorf("failed to parse bot UserID: %w", err)
		}
	}

	if s, ok := m["id"]; ok && s != nil && s != "" {
		if b.ID, err = ulid.ParseStrict(s.(string)); err != nil {
			return fmt.Errorf("failed to parse bot ID [%v]; err: %w", s, err)
		}
	}

	if _, ok := m["target"]; ok {
		b.Target = m["target"].(string)
	}

	if _, ok := m["type"]; ok {
		b.Type = BotType(m["type"].(string))
	}

	if _, ok := m["frequency"]; ok {
		b.Frequency = time.Duration(m["frequency"].(float64))
	}

	if _, ok := m["workedAt"]; ok {
		if b.WorkedAt, err = time.Parse(time.RFC3339, m["workedAt"].(string)); err != nil {
			return fmt.Errorf("failed to parse workedAt: %w", err)
		}
	}

	if val, ok := m["blackList"]; ok {
		for i := 0; i < len(val.([]any)); i++ {
			b.BlackList = append(b.BlackList, val.([]any)[i].(string))
		}
	}

	if val, ok := m["headless"]; ok && val != nil {
		b.Headless = val.(bool)
	}

	if val, ok := m["fingerprint"]; ok {
		if data, err = json.Marshal(val); err != nil {
			log.Warn().Err(fmt.Errorf("failed to marshal fingerprint: %w", err)).Send()
		} else if err = json.Unmarshal(data, &b.Fingerprint); err != nil {
			log.Warn().Err(fmt.Errorf("failed to unmarshal fingerprint: %w", err)).Send()
		}
	}

	return
}

func (b *Bot) String() string {
	byt, _ := json.MarshalIndent(b, "", "\t")
	return string(byt)
}

func (b *Bot) NewBotResult(args ...any) *BotResult {

	m := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		m[args[i].(string)] = args[i+1]
	}

	return &BotResult{
		UserID: b.UserID,
		BotID:  b.ID,
		ID:     NewULID(),
		Type:   b.Type,
		Target: b.Target,
		Data:   m,
	}
}

func (b *Bot) MarshalZerologObject(evt *zerolog.Event) {
	evt.Stringer("userId", b.UserID).
		Stringer("id", b.ID).
		Str("target", b.Target).
		Stringer("type", b.Type).
		Stringer("frequency", b.Frequency).
		Time("workedAt", b.WorkedAt).
		Strs("blackList", b.BlackList).
		Bool("headless", b.Headless).
		Any("fingerprint", b.Fingerprint)
}

func (b *Bot) Key() string {
	return fmt.Sprintf("users/%s/%s/%s.json", b.UserID, b.Type.Plural(), b.Target)
}
