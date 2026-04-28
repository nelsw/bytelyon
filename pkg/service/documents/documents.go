package documents

import (
	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/model"
)

func New(url string) (*model.Document, error) {

	b, err := https.Get(url)
	if err != nil {
		return nil, err
	}

	var doc *model.Document
	if doc, err = model.ParseDocument(string(b)); err != nil {
		return nil, err
	}

	return doc, nil
}
