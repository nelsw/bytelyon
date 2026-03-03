package db

import (
	"context"
	"errors"
	"sync"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/model"
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

var (
	db  *dynamodb.Client
	ctx context.Context

	notFoundEx    *types.ResourceNotFoundException
	tableExistsEx *types.TableAlreadyExistsException
)

func init() {
	ctx = context.Background()
	c := Must(aws.LoadDefaultConfig(ctx))
	db = dynamodb.NewFromConfig(c)

	if !IsReleaseMode() &&
		MigrateTables() {
		Migrate(
			&Email{},
			&BotNews{},
			&BotNewsResult{},
			&Password{},
			&BotSearch{},
			&BotSearchResult{},
			&BotSitemap{},
			&BotSitemapResult{},
			&Token{},
			&User{},
		)
	}
}

// create creates a DynamoDB table.
func create(e Entity) error {

	l := log.With().Str("name", *TableName(e)).Logger()

	l.Trace().Msg("creating table")

	cti := e.GetDesc()
	cti.TableName = TableName(e)
	_, err := db.CreateTable(ctx, &cti)

	if errors.As(err, &tableExistsEx) {
		l.Warn().Msg("table already exists")
		return nil
	}

	if err != nil {
		l.Err(err).Msg("failed to create table")
		return err
	}

	err = dynamodb.NewTableExistsWaiter(db).Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: TableName(e),
	}, 5*time.Minute)

	if err != nil {
		l.Err(err).Msg("failed to wait for table creation")
		return err
	}

	l.Debug().Msg("table created")
	return nil
}

// destroy deletes the DynamoDB table and all of its data.
func destroy(e Entity) error {

	l := log.With().Str("name", *TableName(e)).Logger()

	l.Trace().Msg("deleting table")

	_, err := db.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: TableName(e),
	})

	if errors.As(err, &notFoundEx) {
		log.Warn().Msg("table does not exist")
		return nil
	}

	if err != nil {
		log.Err(err).Msg("failed to delete table")
		return err
	}

	err = dynamodb.NewTableNotExistsWaiter(db).Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: TableName(e),
	}, 5*time.Minute)

	if err != nil {
		log.Warn().Err(err).Msg("failed to wait for table deletion")
		return nil
	}

	log.Debug().Msg("deleted table")
	return nil
}

// Migrate drops and creates the DynamoDB tables defined in the given map.
// Fails fast if app mode is release or migration config equals false.
func Migrate(ee ...Entity) error {

	if IsReleaseMode() || !MigrateTables() {
		return nil
	}

	l := log.With().Int("size", len(ee)).Logger()

	l.Trace().Msg("migrating tables")

	var wg sync.WaitGroup
	for _, e := range ee {
		wg.Go(func() {
			destroy(e)
			create(e)
		})
	}

	l.Trace().Msg("waiting on migrations")

	wg.Wait()

	l.Trace().Msg("migration complete")

	return nil
}

// Wipe removes an item from a DynamoDB table.
func Wipe(e Entity, v any) error {
	l := log.
		With().
		Any("key", v).
		Any("entity", e).
		Str("name", *TableName(e)).
		Logger()

	l.Trace().Msg("deleting item")

	key, err := attributevalue.MarshalMap(v)
	if err != nil {
		l.Err(err).Msg("failed to marshal key")
		return err
	}

	_, err = db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: TableName(e),
		Key:       key,
	})

	if errors.As(err, &notFoundEx) {
		log.Warn().Msg("item does not exist")
		return nil
	}

	if err != nil {
		log.Err(err).Msg("failed to delete item")
		return err
	}

	log.Debug().Msg("item deleted")

	return nil
}

// Save creates a new item, or replaces an old item with a new item.
func Save(e Entity) (err error) {

	l := log.With().Any("entity", e).Str("name", *TableName(e)).Logger()

	l.Trace().Msg("creating item")

	input := dynamodb.PutItemInput{TableName: TableName(e)}
	if input.Item, err = attributevalue.MarshalMap(e); err != nil {
		l.Err(err).Msg("failed to marshal item")
		return err
	}

	if _, err = db.PutItem(ctx, &input); err != nil {
		l.Err(err).Msg("failed to put item")
		return err
	}

	l.Debug().Msg("item created")

	return nil
}

// Find retrieves an item from the DynamoDB table.
func Find[E Entity](a any) (e E, err error) {
	l := log.With().
		Any("any", a).
		Any("entity", e).
		Str("name", *TableName(e)).
		Logger()

	l.Trace().Msg("getting item")

	input := dynamodb.GetItemInput{TableName: TableName(e)}
	if input.Key, err = attributevalue.MarshalMap(a); err != nil {
		l.Err(err).Msg("failed to marshal key")
		return
	}

	var res *dynamodb.GetItemOutput
	res, err = db.GetItem(ctx, &input)

	if err != nil {
		l.Err(err).Msg("failed to get item")
		return
	}

	if res.Item == nil {
		l.Warn().Msg("item not found")
		return
	}

	if err = attributevalue.UnmarshalMap(res.Item, &e); err != nil {
		l.Err(err).Msg("failed to unmarshal item")
		return
	}

	l.Debug().Any("entity", e).Msg("got item")

	return
}

// Query items by the hash key.
// See Bot for a composite key of a hash & range key.
func Query[E Entity](e E, a any) ([]E, error) {

	l := log.With().
		Str("name", *TableName(e)).
		Str("key", KeyName(e)).
		Any("val", a).
		Logger()

	l.Trace().Msg("querying items")

	exp, err := expression.
		NewBuilder().
		WithKeyCondition(expression.Key(KeyName(e)).Equal(expression.Value(a))).
		Build()

	if err != nil {
		l.Err(err).Msg("failed to build expression")
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 TableName(e),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
	}

	var arr []E

	queryPaginator := dynamodb.NewQueryPaginator(db, input)

	var res *dynamodb.QueryOutput
	for queryPaginator.HasMorePages() {

		if res, err = queryPaginator.NextPage(ctx); err != nil {
			l.Err(err).Msg("failed to query")
			return nil, err
		}

		var items []E
		if err = attributevalue.UnmarshalListOfMaps(res.Items, &items); err != nil {
			l.Err(err).Msg("failed to unmarshal items")
			return nil, err
		}
		arr = append(arr, items...)
	}

	l.Debug().Int("size", len(arr)).Msg("queried")

	return arr, nil
}

// Scan is literally a full table scan; don't use this function.
func Scan[E Entity](e E, args ...any) (arr []E, err error) {

	l := log.With().Str("name", *TableName(e)).Logger()

	l.Trace().Msg("scanning items")

	input := &dynamodb.ScanInput{TableName: TableName(e)}

	if len(args) > 0 {
		f := args[0].(string)
		input.FilterExpression = &f
		if input.ExpressionAttributeValues, err = attributevalue.MarshalMap(args[1]); err != nil {
			l.Err(err).Msg("failed to marshal query filter express values")
			return
		}
	}

	var lastEvaluatedKey map[string]types.AttributeValue = nil

	for {

		input.ExclusiveStartKey = lastEvaluatedKey

		var result *dynamodb.ScanOutput
		if result, err = db.Scan(ctx, input); err != nil {
			l.Err(err).Msg("failed to scan item")
			return
		}

		var items []E
		if err = attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
			l.Err(err).Msg("failed to unmarshal items")
			return
		}

		arr = append(arr, items...)
		if result.LastEvaluatedKey == nil {
			break
		}

		lastEvaluatedKey = result.LastEvaluatedKey
	}

	l.Debug().Int("size", len(arr)).Msg("scanned items")

	return
}
