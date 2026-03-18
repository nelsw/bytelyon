package contract

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Tabular interface {
	TableName() *string
}

// Creatable interface provides methods for creating DynamoDB tables.
type Creatable interface {
	// Create returns a CreateTableInput for creating a DynamoDB table.
	Create() *dynamodb.CreateTableInput
}

// Gettable interface provides methods for retrieving DynamoDB items.
type Gettable interface {
	// Get returns a GetItemInput for retrieving a DynamoDB item.
	Get() *dynamodb.GetItemInput
}

// Puttable interface provides methods for creating or updating DynamoDB items.
type Puttable interface {
	// Put returns a PutItemInput for creating or updating a DynamoDB item.
	Put() *dynamodb.PutItemInput
}

// Queryable interface provides methods for querying DynamoDB tables.
type Queryable interface {
	// Query returns a QueryInput for querying a DynamoDB table.
	Query() *dynamodb.QueryInput
}

// Scannable interface provides methods for scanning DynamoDB tables.
type Scannable interface {
	// Scan returns a ScanInput for scanning a DynamoDB table.
	Scan() *dynamodb.ScanInput
}
