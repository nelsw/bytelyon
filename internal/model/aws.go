package model

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWS struct {
	aws.Credentials
	Bucket string `gorm:"size:255"`
	Region string `gorm:"size:255"`
}
