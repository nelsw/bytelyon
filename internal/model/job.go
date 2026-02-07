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
	Type      JobType        `gorm:"index:idx_job_type_target_deleted,unique"`
	Target    string         `gorm:"index:idx_job_type_target_deleted,unique"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_job_type_target_deleted,unique"`
	CreatedAt int
	UpdatedAt int `gorm:"<-:false"`
	Frequency time.Duration
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

	return time.Unix(int64(j.UpdatedAt), 0).Add(j.Frequency).Before(time.Now())
}
