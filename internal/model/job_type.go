package model

import (
	"database/sql/driver"
	"errors"
	"regexp"
)

var (
	validationRegex = regexp.MustCompile(`^(search|article|sitemap)$`)
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

func NewJobType(s string) (JobType, error) {
	if !validationRegex.MatchString(s) {
		return "", errors.New("invalid job type, must be one of: search, article, sitemap")
	}
	return JobType(s), nil
}

func (t JobType) String() string {
	return string(t)
}
