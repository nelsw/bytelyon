package db

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var notFoundEx *types.ResourceNotFoundException

func DeleteItem(ctx context.Context, c *dynamodb.Client, input *dynamodb.DeleteItemInput) error {
	_, err := c.DeleteItem(ctx, input)
	if err == nil || errors.Is(err, notFoundEx) {
		return nil
	}
	return err
}

func GetItem(ctx context.Context, c *dynamodb.Client, input *dynamodb.GetItemInput) (map[string]types.AttributeValue, error) {
	out, err := c.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}
	return out.Item, nil
}

func PutItem(ctx context.Context, c *dynamodb.Client, input *dynamodb.PutItemInput) error {
	_, err := c.PutItem(ctx, input)
	return err
}

func QueryItems(ctx context.Context, c *dynamodb.Client, input *dynamodb.QueryInput) (arr []map[string]types.AttributeValue, err error) {
	var res *dynamodb.QueryOutput
	for paginator := dynamodb.NewQueryPaginator(c, input); paginator.HasMorePages(); {
		if res, err = paginator.NextPage(ctx); err != nil {
			return
		}
		for _, item := range res.Items {
			arr = append(arr, item)
		}
	}
	return
}
