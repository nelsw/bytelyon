package news

import (
	"fmt"

	"github.com/nelsw/bytelyon/pkg/article"
	"github.com/oklog/ulid/v2"
)

// Model contains topical news articles that belong to a user.
type Model struct {

	// Entries is a map of article URLs to model.Page keys.
	Entries map[string]ulid.ULID `json:"entries"`

	// Topic is the news topic, defined by the model.Bot target.
	Topic string `json:"-"`

	// UserID is the user ID of the user who is requesting the news.
	UserID ulid.ULID `json:"-"`

	// Articles is a list of articles that belong to the topic.
	Articles article.Models `json:"articles,omitempty"`
}

func New(userID ulid.ULID, topic string) *Model {
	return &Model{
		Entries: make(map[string]ulid.ULID),
		Topic:   topic,
		UserID:  userID,
	}
}

func (m *Model) Merge(other *Model) {
	for k, v := range other.Entries {
		if _, ok := m.Entries[k]; !ok {
			m.Entries[k] = v
		}
	}
}

func (m *Model) Key() string {
	return fmt.Sprintf("users/%s/news/%s.json", m.UserID, m.Topic)
}
