package db

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/aws"
	. "github.com/nelsw/bytelyon/pkg/contract"
	"github.com/rs/zerolog/log"
)

var notFoundEx *types.ResourceNotFoundException
var tableExistsEx *types.TableAlreadyExistsException

var ctx = context.Background()
var db = aws.DB()

// create creates a DynamoDB table.
func Create(t Creatable) error {

	log.Trace().Any("table", t.Create().TableName).Msg("creating table")

	_, err := db.CreateTable(ctx, t.Create())

	if errors.As(err, &tableExistsEx) {
		log.Warn().Err(err).Any("table", t).Msg("table already exists")
		return nil
	}

	if err != nil {
		log.Err(err).Any("table", t).Msg("failed to create table")
		return err
	}

	err = dynamodb.
		NewTableExistsWaiter(db).
		Wait(ctx, &dynamodb.DescribeTableInput{TableName: t.Create().TableName}, 5*time.Minute)

	if err != nil {
		log.Err(err).Any("table", t).Msg("failed to wait for table creation")
		return err
	}

	log.Debug().Any("table", t.Create().TableName).Msg("table created")

	return nil
}

// Drop deletes the DynamoDB table and all of its data.
func Drop(tableName *string) error {

	log.Trace().Any("table", tableName).Msg("deleting table")

	_, err := db.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: tableName})

	if errors.As(err, &notFoundEx) {
		log.Warn().Any("table", tableName).Msg("table does not exist")
		return nil
	}

	if err != nil {
		log.Err(err).Any("table", tableName).Msg("failed to delete table")
		return err
	}

	err = dynamodb.
		NewTableNotExistsWaiter(db).
		Wait(ctx, &dynamodb.DescribeTableInput{TableName: tableName}, 5*time.Minute)

	if err != nil {
		log.Warn().Err(err).Any("table", tableName).Msg("failed to wait for table deletion")
		return nil
	}

	log.Debug().Any("table", tableName).Msg("deleted table")
	return nil
}

// Migrate drops and creates the DynamoDB tables defined in the given map.
// Fails fast if app mode is release or migration config equals false.
func Migrate(tt ...Creatable) error {

	//if os.Getenv("MIGRATE_TABLES") != "true" {
	//	return nil
	//}

	l := log.With().Int("size", len(tt)).Logger()

	for _, e := range tt {
		//Drop(e.Create().tableName)
		Create(e)
	}

	l.Trace().Msg("waiting on migrations")

	l.Trace().Msg("migration complete")

	return nil
}

// Delete removes an item from a DynamoDB table.
func Delete(t Gettable) error {
	l := log.With().Any("table", t.Get().TableName).Logger()

	l.Trace().Msg("deleting item")

	_, err := db.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: t.Get().TableName, Key: t.Get().Key})

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

// Put creates a new item or replaces an old item with a new item.
func Put(t Puttable) error {

	l := log.With().Any("table", t.Put().TableName).Logger()

	l.Trace().Msg("creating item")

	if _, err := db.PutItem(ctx, t.Put()); err != nil {
		l.Err(err).Msg("failed to put item")
		return err
	}

	l.Debug().Msg("item created")

	return nil
}

// Get retrieves an item from the DynamoDB table.
func Get[T any](t Gettable) (out T, err error) {
	l := log.With().Any("t", t.Get().TableName).Logger()

	l.Trace().Msg("get item")

	var res *dynamodb.GetItemOutput
	if res, err = db.GetItem(ctx, t.Get()); err != nil {
		l.Err(err).Msg("get failed!")
		return
	}
	err = attributevalue.UnmarshalMap(res.Item, &out)
	if err != nil {
		l.Err(err).Msg("failed to unmarshal item")
	}
	l.Debug().Any("table", &out).Msg("got item")
	return
}

// Query items by the hash key.
// See Bot for a composite key of a hash and range key.
func Query[T any](t Queryable) (out []T, err error) {

	l := log.With().Any("t", t.Query().TableName).Logger()

	l.Trace().Msg("querying items")

	var res *dynamodb.QueryOutput

	paginator := dynamodb.NewQueryPaginator(db, t.Query())
	for paginator.HasMorePages() {

		if res, err = paginator.NextPage(ctx); err != nil {
			l.Err(err).Msg("failed to query")
			return nil, err
		}

		if err = attributevalue.UnmarshalListOfMaps(res.Items, &out); err != nil {
			l.Err(err).Msg("failed to unmarshal items")
			return nil, err
		}
		out = append(out, out...)
	}

	l.Debug().Int("size", len(out)).Msg("queried")

	return out, nil
}

// Scan is literally a full table scan; don't use this function.
func Scan[T any](t Scannable) ([]T, error) {

	l := log.With().Any("t", t.Scan().TableName).Logger()

	l.Trace().Msg("scanning items")

	input := t.Scan()

	var lastEvaluatedKey map[string]types.AttributeValue = nil
	var out []T
	for {

		input.ExclusiveStartKey = lastEvaluatedKey

		result, err := db.Scan(ctx, input)
		if err != nil {
			l.Err(err).Msg("failed to scan item")
			return nil, err
		}

		if err = attributevalue.UnmarshalListOfMaps(result.Items, &out); err != nil {
			l.Err(err).Msg("failed to unmarshal items")
			return nil, err
		}
		out = append(out, out...)

		if result.LastEvaluatedKey == nil {
			break
		}

		lastEvaluatedKey = result.LastEvaluatedKey
	}

	l.Debug().Int("size", len(out)).Msg("scanned items")

	return out, nil
}
