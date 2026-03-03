package model

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/internal/config"

	. "github.com/nelsw/bytelyon/internal/util"
)

type Entity interface {
	GetDesc() dynamodb.CreateTableInput
}

func TableName(e Entity) *string {
	n := "ByteLyon_"
	if config.IsDebugMode() {
		n += "Debug_"
	} else if config.IsTestMode() {
		n += "Test_"
	}
	n += strings.Join(SplitByCase(Name(e)), "_")
	return &n
}

func KeyName(e Entity) string {
	d := e.GetDesc()
	for _, k := range d.KeySchema {
		if k.KeyType == types.KeyTypeHash {
			return *k.AttributeName
		}
	}
	return ""
}
