package model

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"gorm.io/gorm"
)

type Settings struct {
	gorm.Model
	AWS
}

func (s *Settings) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return s.Credentials, nil
}
