package contract

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/model"
)

type News interface {
	GetBody() []string
	GetDescription() string
	GetImage() model.Image
	GetPublishedAt() time.Time
	GetSource() string
	GetTitle() string
	GetURL() string
}
