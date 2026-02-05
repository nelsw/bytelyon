package model

import (
	"database/sql/driver"
	"errors"
)

type JobType string

const (
	SearchType  JobType = "search"
	SitemapType JobType = "sitemap"
	ArticleType JobType = "article"
)

func (t *JobType) Scan(src any) error {
	if src == nil {
		return errors.New("job type is nil")
	}
	str, ok := src.(string)
	if !ok {
		return errors.New("invalid job type")
	}
	*t = JobType(str)
	return nil
}

func (t *JobType) Value() (driver.Value, error) {
	return string(*t), nil
}
