package entity

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type News struct {
	Topic string

	pages map[string]ulid.ULID

	articles map[string]model.Article

	userID    ulid.ULID
	workedAt  time.Time
	blackList []string
}

func (e *News) key() string {
	return fmt.Sprintf("users/%s/news/%s.json", e.userID, e.Topic)
}

func (e *News) Save() {
	s3.PutPrivateObject(e.key(), util.JSON(e))
}

func (e *News) From(bot *model.Bot) *News {
	if f := e.Find(bot.UserID, bot.Target); f != nil {
		return f
	}
	return e.Create(bot)
}

func (e *News) Create(bot *model.Bot) *News {
	n := &News{
		pages:     make(map[string]ulid.ULID),
		articles:  make(map[string]model.Article),
		Topic:     bot.Target,
		userID:    bot.UserID,
		workedAt:  bot.WorkedAt,
		blackList: bot.BlackList,
	}
	n.Save()
	return n
}

func (e *News) Delete(userID ulid.ULID, topic string, pageID ...ulid.ULID) {
	e.userID = userID
	e.Topic = topic
	if len(pageID) == 0 {
		s3.DeletePrivateObject(e.key())
		return
	}
	e.Find(userID, topic)
}

func (e *News) Find(userID ulid.ULID, topic string) *News {

	e.userID = userID
	e.Topic = topic

	if out, err := s3.GetPrivateObject(e.key()); err != nil {
		return nil
	} else if err = json.Unmarshal(out, e); err != nil {
		return nil
	}

	return e
}

func (e *News) Add(p *Page, pubDate time.Time, source, description string) {

	if p == nil {
		return
	} else if _, ok := e.pages[p.URL]; ok {
		return
	} else if !e.workedAt.IsZero() && pubDate.Before(e.workedAt) {
		return
	}

	p.Save()

	blackMap := make(map[string]bool)
	for _, word := range e.blackList {
		blackMap[word] = true
	}

	a := p.MakeArticle(pubDate, source, description)
	for _, word := range a.Words() {
		if _, ok := blackMap[word]; ok {
			return
		}
	}

	e.pages[p.URL] = p.ID
	e.articles[p.URL] = a

	e.Save()
}

func (e *News) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"articles": slices.SortedFunc(maps.Values(e.articles), func(a, z model.Article) int {
			return a.PublishedAt.Compare(z.PublishedAt)
		}),
		"pages": e.pages,
		"topic": e.Topic,
	})
}

func (e *News) UnmarshalJSON(b []byte) error {

	var m map[string]any

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	e.articles = m["articles"].(map[string]model.Article)
	e.pages = m["pages"].(map[string]ulid.ULID)
	e.Topic = m["topic"].(string)

	return nil
}
