package contract

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type Entity interface {
	Desc() *dynamodb.CreateTableInput
	Name() string
	Key() map[string]any
	Validate() error
}
