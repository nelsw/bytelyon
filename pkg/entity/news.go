package entity

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

type News struct {
	*model.SyncMap[string, ulid.ULID]
	Topic   string
	UserID  ulid.ULID
	After   time.Time
	Exclude map[string]bool
}

func NewNews(userID ulid.ULID, topic string) *News {
	return &News{
		Topic:   topic,
		SyncMap: model.NewSyncMap[string, ulid.ULID](),
		UserID:  userID,
	}
}

func (e *News) Key() string {
	return fmt.Sprintf("users/%s/news/%s.json", e.UserID, e.Topic)
}

func (e *News) MarshalJSON() ([]byte, error) {
	return json.Marshal(slices.SortedFunc(maps.Values(e.Map), func(a, z ulid.ULID) int { return a.Compare(z) }))
}

func (e *News) UnmarshalJSON(b []byte) error {
	var m map[string]ulid.ULID
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	e.SyncMap = model.NewSyncMap[string, ulid.ULID](m)

	e.After = time.Time{}
	e.Exclude = make(map[string]bool)
	for k, v := range m {
		e.Exclude[k] = true
		if v.Timestamp().After(e.After) {
			e.After = v.Timestamp()
		}
	}

	return nil
}
