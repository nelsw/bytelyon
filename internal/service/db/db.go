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
	c, err := aws.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	db = dynamodb.NewFromConfig(c)
}

// Make creates a DynamoDB table.
func Make(name string, desc *dynamodb.CreateTableInput) error {

	log.Trace().Str("name", name).Msg("creating table")

	_, err := db.CreateTable(ctx, desc)

	if errors.As(err, &tableExistsEx) {
		log.Warn().Str("name", name).Msg("table already exists")
		return nil
	}

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to create table")
		return err
	}

	err = dynamodb.NewTableExistsWaiter(db).Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: Ptr(name),
	}, 5*time.Minute)

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to wait for table creation")
		return err
	}

	log.Debug().Str("name", name).Msg("table created")
	return nil
}

// Drop deletes the DynamoDB table and all of its data.
func Drop(name string) error {

	log.Trace().Str("name", name).Msg("deleting table")

	_, err := db.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: Ptr(name),
	})

	if errors.As(err, &notFoundEx) {
		log.Warn().Str("name", name).Msg("table does not exist")
		return nil
	}

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to delete table")
		return err
	}

	err = dynamodb.NewTableNotExistsWaiter(db).Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: Ptr(name),
	}, 5*time.Minute)

	if err != nil {
		log.Warn().Err(err).Str("name", name).Msg("failed to wait for table deletion")
		return nil
	}

	log.Debug().Str("name", name).Msg("deleted table")
	return nil
}

// Migrate drops and creates the DynamoDB tables defined in the given map.
// Fails fast if app mode is release or migration config equals false.
func Migrate(arr ...*dynamodb.CreateTableInput) {

	if IsReleaseMode() || !MigrateTables() {
		return
	}

	l := log.With().Int("size", len(arr)).Logger()

	l.Trace().Msg("migrating tables")

	var wg sync.WaitGroup
	for _, a := range arr {
		wg.Go(func() {
			Drop(*a.TableName)
			Make(*a.TableName, a)
		})
	}

	l.Trace().Msg("waiting on migrations")

	wg.Wait()

	l.Trace().Msg("migration complete")
}

// Wipe removes an item from a DynamoDB table.
func Wipe(a, v any) error {
	l := log.With().Str("name", *TableName(a)).Logger()

	l.Trace().Msg("deleting item")

	key, err := attributevalue.MarshalMap(v)
	if err != nil {
		l.Err(err).Msg("failed to marshal key")
		return err
	}

	_, err = db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: TableName(a),
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
func Save(a any) (err error) {

	l := log.With().Str("name", *TableName(a)).Logger()

	l.Trace().Msg("creating item")

	input := dynamodb.PutItemInput{TableName: TableName(a)}
	if input.Item, err = attributevalue.MarshalMap(a); err != nil {
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
func Find[T any](a, v any) (t T, err error) {
	l := log.With().Str("name", *TableName(a)).Logger()

	l.Trace().Msg("getting item")

	input := dynamodb.GetItemInput{TableName: TableName(a)}
	if input.Key, err = attributevalue.MarshalMap(v); err != nil {
		log.Err(err).Msg("failed to marshal key")
		return
	}

	log.Trace().Msg("getting item")

	var res *dynamodb.GetItemOutput
	res, err = db.GetItem(ctx, &input)

	if err != nil {
		log.Err(err).Msg("failed to get item")
		return
	}

	if res.Item == nil {
		log.Warn().Msg("item not found")
		return
	}

	if err = attributevalue.UnmarshalMap(res.Item, &t); err != nil {
		log.Err(err).Msg("failed to unmarshal item")
		return
	}

	log.Debug().Any("item", t).Msg("got item")

	return
}

// Query items by the hash key.
// See model.Bot for a composite key of a hash & range key.
func Query[T any](t T, k string, v any) ([]T, error) {

	l := log.With().
		Str("name", *TableName(t)).
		Str("key", k).
		Any("val", v).
		Logger()

	l.Trace().Msg("querying items")

	exp, err := expression.
		NewBuilder().
		WithKeyCondition(expression.Key(k).Equal(expression.Value(v))).
		Build()

	if err != nil {
		l.Err(err).Msg("failed to build expression")
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 TableName(t),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
	}

	var arr []T

	queryPaginator := dynamodb.NewQueryPaginator(db, input)

	var res *dynamodb.QueryOutput
	for queryPaginator.HasMorePages() {

		if res, err = queryPaginator.NextPage(ctx); err != nil {
			l.Err(err).Msg("failed to query")
			return nil, err
		}

		var items []T
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
func Scan[T any](t T, args ...any) (arr []T, err error) {

	l := log.With().Str("name", *TableName(t)).Logger()

	l.Trace().Msg("scanning items")

	input := &dynamodb.ScanInput{TableName: TableName(t)}

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

		var items []T
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
