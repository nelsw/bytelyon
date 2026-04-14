package documents

import (
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/model"
)

func New(url string) (*model.Document, error) {

	b, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	var doc *model.Document
	if doc, err = model.ParseDocument(string(b)); err != nil {
		return nil, err
	}

	return doc, nil
}
