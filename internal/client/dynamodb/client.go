package client

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/rs/zerolog/log"
)

var (
	NotFoundEx    *types.ResourceNotFoundException
	TableExistsEx *types.TableAlreadyExistsException
)

// TableExists determines whether a DynamoDB table exists.
func TableExists(ctx context.Context, c *dynamodb.Client, name string) (bool, error) {
	log.Trace().Str("name", name).Msg("checking if table exists")

	_, err := c.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &name,
	})

	if errors.As(err, &NotFoundEx) {
		log.Debug().Str("name", name).Msg("table does not exist")
		return false, nil
	}

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to determine table existence")
		return false, err
	}

	log.Debug().Str("name", name).Msg("table exists")
	return true, nil
}

// CreateTable creates a DynamoDB table.
func CreateTable(ctx context.Context, c *dynamodb.Client, input *dynamodb.CreateTableInput) error {
	log.Trace().Str("name", *input.TableName).Msg("creating table")

	_, err := c.CreateTable(ctx, input)

	if errors.As(err, &TableExistsEx) {
		log.Warn().Str("name", *input.TableName).Msg("table already exists")
		return nil
	}

	if err != nil {
		log.Err(err).Str("name", *input.TableName).Msg("failed to create table")
		return err
	}

	err = dynamodb.NewTableExistsWaiter(c).Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: input.TableName,
	}, 5*time.Minute)

	if err != nil {
		log.Err(err).Str("name", *input.TableName).Msg("failed to wait for table creation")
		return err
	}

	log.Debug().Str("name", *input.TableName).Msg("table created")
	return nil
}

// DeleteTable deletes the DynamoDB table and all of its data.
func DeleteTable(ctx context.Context, c *dynamodb.Client, name string) error {
	log.Trace().Str("name", name).Msg("deleting table")

	_, err := c.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: &name,
	})

	if errors.As(err, &NotFoundEx) {
		log.Warn().Str("name", name).Msg("table does not exist")
		return nil
	}

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to delete table")
		return err
	}

	err = dynamodb.NewTableNotExistsWaiter(c).Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: &name,
	}, 5*time.Minute)

	if err != nil {
		log.Warn().Err(err).Str("name", name).Msg("failed to wait for table deletion")
		return nil
	}

	log.Debug().Str("name", name).Msg("deleted table")
	return nil
}

// ListTables lists the DynamoDB table names for the current account.
func ListTables(ctx context.Context, c *dynamodb.Client) ([]string, error) {
	log.Trace().Msg("listing tables")

	var names []string
	var output *dynamodb.ListTablesOutput
	var err error

	tablePaginator := dynamodb.NewListTablesPaginator(c, &dynamodb.ListTablesInput{})

	for tablePaginator.HasMorePages() {
		if output, err = tablePaginator.NextPage(ctx); err != nil {
			log.Err(err).Msg("failed to list tables")
			return nil, err
		}
		names = append(names, output.TableNames...)
	}

	log.Debug().Strs("names", names).Msg("listed tables")

	return names, nil
}

// PutItem creates a new item, or replaces an old item with a new item.
func PutItem(ctx context.Context, c *dynamodb.Client, name string, a any) error {

	log.Trace().Str("name", name).Msg("creating item")

	item, err := attributevalue.MarshalMap(a)
	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to marshal item")
		return err
	}

	_, err = c.PutItem(ctx, &dynamodb.PutItemInput{
		Item:      item,
		TableName: &name,
	})

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to put item")
		return err
	}

	log.Debug().Str("name", name).Msg("item created")

	return nil
}

// DeleteItem removes an item from a DynamoDB table.
func DeleteItem(ctx context.Context, c *dynamodb.Client, name string, a any) error {

	log.Trace().Str("name", name).Msg("deleting item")

	key, err := attributevalue.MarshalMap(a)
	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to marshal key")
		return err
	}

	_, err = c.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &name,
		Key:       key,
	})

	if errors.As(err, &NotFoundEx) {
		log.Warn().Str("name", name).Msg("item does not exist")
		return nil
	}

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to delete item")
		return err
	}

	log.Debug().Str("name", name).Msg("item deleted")

	return nil
}

// GetItem retrieves an item from the DynamoDB table.
func GetItem[T any](ctx context.Context, c *dynamodb.Client, name string, a any) (t T, err error) {

	log.Trace().Str("name", name).Msg("getting item")

	var key map[string]types.AttributeValue
	if key, err = attributevalue.MarshalMap(a); err != nil {
		log.Err(err).Str("name", name).Msg("failed to marshal key")
		return
	}

	var res *dynamodb.GetItemOutput
	res, err = c.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &name,
		Key:       key,
	})

	if err != nil {
		log.Err(err).Str("name", name).Msg("failed to get item")
		return
	}

	if res.Item == nil {
		log.Warn().Str("name", name).Msg("item not found")
		err = NotFoundEx
		return
	}

	if err = attributevalue.UnmarshalMap(res.Item, &t); err != nil {
		log.Err(err).Str("name", name).Msg("failed to unmarshal item")
		return
	}

	log.Debug().Str("name", name).Any("item", t).Msg("got item")

	return
}

// QueryByID gets all items in the DynamoDB table by the hash key.
func QueryByID[T any](ctx context.Context, c *dynamodb.Client, name, key string, val uuid.UUID) ([]T, error) {

	l := log.With().
		Str("name", name).
		Str("key", key).
		Stringer("val", val).
		Logger()

	l.Trace().Msg("querying")

	expr, err := expression.
		NewBuilder().
		WithKeyCondition(expression.Key(key).Equal(expression.Value(val))).
		Build()

	if err != nil {
		l.Err(err).Msg("failed to build expression")
		return nil, err
	}

	queryPaginator := dynamodb.NewQueryPaginator(c, &dynamodb.QueryInput{
		TableName:                 &name,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	var response *dynamodb.QueryOutput
	var arr []T
	for queryPaginator.HasMorePages() {

		if response, err = queryPaginator.NextPage(ctx); err != nil {
			l.Err(err).Msg("failed to query")
			return nil, err
		}

		var tt []T
		if err = attributevalue.UnmarshalListOfMaps(response.Items, &tt); err != nil {
			l.Err(err).Msg("failed to unmarshal items")
			return nil, err
		}
		arr = append(arr, tt...)
	}

	l.Debug().Int("size", len(arr)).Msg("queried")

	return arr, err
}

// New returns a new DynamoDB client with the given Region, AccessKeyID, and SecretAccessKey.
func New() *dynamodb.Client {
	return dynamodb.NewFromConfig(config.AWS())
}
