package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/aws"
)

// Gettable interface provides methods for retrieving DynamoDB items.
type Gettable interface {
	// Get returns a GetItemInput for retrieving a DynamoDB item.
	Get() *dynamodb.GetItemInput
}

// Queryable interface provides methods for querying DynamoDB tables.
type Queryable interface {
	// Query returns a QueryInput for querying a DynamoDB table.
	Query() *dynamodb.QueryInput
}

// Delete removes an item from a DynamoDB table.
func Delete(t Gettable) error {
	return DeleteItem(context.Background(), aws.DB(), &dynamodb.DeleteItemInput{
		TableName: t.Get().TableName,
		Key:       t.Get().Key,
	})
}

// Put creates a new item or replaces an old item with a new item.
func Put(t Gettable) error {
	item, err := attributevalue.MarshalMap(&t)
	if err != nil {
		return err
	}
	return PutItem(context.Background(), aws.DB(), &dynamodb.PutItemInput{
		TableName: t.Get().TableName,
		Item:      item,
	})
}

// Get retrieves an item from the DynamoDB table.
func Get[T Gettable](t T) (T, error) {
	item, err := GetItem(context.Background(), aws.DB(), t.Get())
	if err == nil {
		err = attributevalue.UnmarshalMap(item, &t)
	}
	return t, err
}

// Query items by the hash key.
func Query[T Queryable](t T) (out []T, err error) {
	var items []map[string]types.AttributeValue
	if items, err = QueryItems(context.Background(), aws.DB(), t.Query()); err == nil {
		err = attributevalue.UnmarshalListOfMaps(items, &out)
	}
	return
}
