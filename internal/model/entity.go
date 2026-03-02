package model

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nelsw/bytelyon/internal/config"

	. "github.com/nelsw/bytelyon/internal/util"
)

type Entity interface {
	Desc() dynamodb.CreateTableInput
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
