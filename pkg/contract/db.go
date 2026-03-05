package contract

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Creatable interface {
	Create() *dynamodb.CreateTableInput
}

type Gettable interface {
	Get() *dynamodb.GetItemInput
}

type Puttable interface {
	Put() *dynamodb.PutItemInput
}

type Queryable interface {
	Query() *dynamodb.QueryInput
}

type Scannable interface {
	Scan() *dynamodb.ScanInput
}
