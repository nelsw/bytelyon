package service

import (
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/worker"
	"gorm.io/gorm"
)

type JobService interface {
	List() ([]*model.Job, error)
	Save(*model.Job) error
	Delete(uint) error
}

type jobService struct {
	*gorm.DB
}

func NewJobService(db *gorm.DB) JobService {
	return &jobService{db}
}

func (j jobService) List() (arr []*model.Job, err error) {
	err = j.Find(&arr).Error
	return
}

func (j jobService) Save(job *model.Job) error {
	err := j.DB.Save(job).Error
	if err != nil {
		return err
	}
	if job.CreatedAt == job.UpdatedAt {
		worker.New(j.DB, job).Work()
	}
	return nil
}

func (j jobService) Delete(id uint) error {
	return j.DB.Where("id = ?", id).Delete(&model.Job{}).Error
}
