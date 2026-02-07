package model

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

var (
	urlValidationRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
)

type Job struct {
	gorm.Model
	Type      JobType
	Frequency time.Duration
	Target    string   `gorm:"type:varchar(255)"`
	BlackList []string `gorm:"serializer:json"`
}

func (j *Job) Ignore() map[string]bool {
	m := map[string]bool{}
	for _, s := range j.BlackList {
		m[s] = true
	}
	return m
}

func (j *Job) Validate() error {
	if j.Type == ArticleType {
		if err := urlValidationRegex.MatchString(j.Target); err {
			return fmt.Errorf("bad url, must begin with https://")
		}
	}
	return nil
}

func (j *Job) ReadyToWork() bool {
	if j.CreatedAt == j.UpdatedAt {
		return true
	}
	if j.Frequency == 0 {
		return false
	}
	return j.UpdatedAt.Add(j.Frequency).Before(time.Now())
}
