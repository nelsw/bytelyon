package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type NewsBot struct{ Bot }

func (b NewsBot) Desc() dynamodb.CreateTableInput { return b.Bot.desc() }
