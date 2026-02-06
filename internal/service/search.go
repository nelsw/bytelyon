package service

import (
	"github.com/nelsw/bytelyon/internal/model"
	"gorm.io/gorm"
)

type SearchService interface {
	Delete(uint) error
}

type searchService struct {
	*gorm.DB
}

func NewSearchService(db *gorm.DB) SearchService {
	return &searchService{db}
}

func (s *searchService) Delete(id uint) error {
	var a model.SearchPage
	a.ID = id
	return s.Model(&a).Association("Pages").Unscoped().Clear()
}
