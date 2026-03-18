package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/aws"
	client "github.com/nelsw/bytelyon/pkg/client/dynamodb"
	. "github.com/nelsw/bytelyon/pkg/contract"
	"github.com/rs/zerolog/log"
)

var ctx = context.Background()
var db = aws.DB()

// createTable creates a DynamoDB table.
func createTable(t Creatable) error {
	log.Debug().Any("table", t.Create().TableName).Msg("create")
	err := client.CreateTable(ctx, db, t.Create())
	log.Err(err).Any("table", t.Create().TableName).Msg("create")
	return err
}

// deleteTable deletes a DynamoDB table.
func deleteTable(t Creatable) error {
	log.Debug().Any("table", t.Create().TableName).Msg("delete")
	err := client.DeleteTable(ctx, db, &dynamodb.DeleteTableInput{TableName: t.Create().TableName})
	log.Err(err).Any("table", t.Create().TableName).Msg("delete")
	return err
}

// Create creates a DynamoDB table.
func Create(t Creatable) error {
	log.Debug().Any("table", t.Create().TableName).Msg("create")
	err := client.CreateTable(ctx, db, t.Create())
	log.Err(err).Any("table", t.Create().TableName).Msg("create")
	return err
}

// Migrate drops and creates the DynamoDB tables defined in the given map.
// Fails fast if app mode is release or migration config equals false.
func Migrate(tt ...Creatable) {
	log.Debug().Int("size", len(tt)).Msg("migrate")
	for _, t := range tt {
		deleteTable(t)
		createTable(t)
	}
	log.Info().Int("size", len(tt)).Msg("migrate")
}

// Delete removes an item from a DynamoDB table.
func Delete(t Gettable) error {
	log.Debug().Any("table", t.Get().TableName).Msg("delete")
	err := client.DeleteItem(ctx, db, &dynamodb.DeleteItemInput{TableName: t.Get().TableName, Key: t.Get().Key})
	log.Err(err).Any("table", t.Get().TableName).Msg("delete")
	return nil
}

// Put creates a new item or replaces an old item with a new item.
func Put(t Puttable) error {
	log.Debug().Any("table", t.Put().TableName).Msg("put")
	err := client.PutItem(context.Background(), db, t.Put())
	log.Err(err).Any("table", t.Put().TableName).Msg("put")
	return err
}

func PutItem(t Gettable) error {
	log.Debug().Any("table", t.Get().TableName).Msg("put")
	item, err := attributevalue.MarshalMap(&t)
	if err == nil {
		err = client.PutItem(context.Background(), db, &dynamodb.PutItemInput{
			TableName: t.Get().TableName,
			Item:      item,
		})
	}
	log.Err(err).Any("table", t.Get().TableName).Msg("put")
	return err
}

// Get retrieves an item from the DynamoDB table.
func Get[T Gettable](t T) (T, error) {
	log.Debug().Any("table", t.Get().TableName).Msg("get")
	item, err := client.GetItem(ctx, db, t.Get())
	if err == nil {
		err = attributevalue.UnmarshalMap(item, &t)
	}
	log.Err(err).Any("table", t.Get().TableName).Msg("get")
	return t, err
}

// Query items by the hash key.
// See Bot for a composite key of a hash and range key.
func Query[T Queryable](t T) (out []T, err error) {
	log.Debug().Any("table", t.Query().TableName).Msg("query")
	var items []map[string]types.AttributeValue
	if items, err = client.QueryItems(ctx, db, t.Query()); err == nil {
		err = attributevalue.UnmarshalListOfMaps(items, &out)
	}
	log.Err(err).Any("table", t.Query().TableName).Int("size", len(out)).Msg("query")
	return
}

// Scan is literally a full table scan; don't use this function.
func Scan[T Scannable](t T) (out []T, err error) {
	log.Debug().Any("table", t.Scan().TableName).Msg("scan")
	var items []map[string]types.AttributeValue
	if items, err = client.ScanItems(ctx, db, t.Scan()); err == nil {
		err = attributevalue.UnmarshalListOfMaps(items, &out)
	}
	log.Err(err).Any("table", t.Scan().TableName).Int("size", len(out)).Msg("scan")
	return
}
