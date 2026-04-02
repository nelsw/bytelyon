package client

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
)

var notFoundEx *types.ResourceNotFoundException

func DeleteItem(ctx context.Context, c *dynamodb.Client, input *dynamodb.DeleteItemInput) error {
	input.TableName = tableName(input.TableName)
	_, err := c.DeleteItem(ctx, input)
	if err == nil || errors.Is(err, notFoundEx) {
		return nil
	}
	return err
}

func GetItem(ctx context.Context, c *dynamodb.Client, input *dynamodb.GetItemInput) (map[string]types.AttributeValue, error) {
	input.TableName = tableName(input.TableName)
	out, err := c.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}
	return out.Item, nil
}

func PutItem(ctx context.Context, c *dynamodb.Client, input *dynamodb.PutItemInput) error {
	input.TableName = tableName(input.TableName)
	_, err := c.PutItem(ctx, input)
	return err
}

func QueryItems(ctx context.Context, c *dynamodb.Client, input *dynamodb.QueryInput) (arr []map[string]types.AttributeValue, err error) {
	input.TableName = tableName(input.TableName)
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

func ScanItems(ctx context.Context, c *dynamodb.Client, input *dynamodb.ScanInput) ([]map[string]types.AttributeValue, error) {
	input.TableName = tableName(input.TableName)
	var lastEvaluatedKey map[string]types.AttributeValue = nil
	var out []map[string]types.AttributeValue
	for {
		input.ExclusiveStartKey = lastEvaluatedKey

		result, err := c.Scan(ctx, input)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan item")
			return nil, err
		}

		out = append(out, result.Items...)

		if result.LastEvaluatedKey == nil {
			break
		}

		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return out, nil
}

func tableName(ptr *string) *string {
	val := "ByteLyon_" + *ptr
	if os.Getenv("MODE") != "release" {
		val = os.Getenv("MODE") + "_" + val
	}
	return &val
}
