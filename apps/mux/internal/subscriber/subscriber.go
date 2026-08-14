package subscriber

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/internal/client"
	"github.com/nelsw/bytelyon/apps/mux/internal/db"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/nelsw/bytelyon/apps/mux/internal/s3"
	"github.com/nelsw/bytelyon/apps/mux/internal/trait"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type channel int

const (
	bots channel = iota
	news
	pages
	searches
	sitemaps
)

var channels = []string{"bots", "news", "pages", "searches", "sitemaps"}

func (c channel) String() string { return channels[c] }

type Subscriber struct {
	rps []*redis.PubSub
	wg  sync.WaitGroup
}

func Start() *Subscriber {

	s := &Subscriber{}

	for i, name := range channels {
		s.rps = append(s.rps, db.Sub(name))
		s.wg.Go(func() {
			for msg := range s.rps[i].Channel() {
				s.do(channel(i), []byte(msg.Payload))
			}
		})
	}

	time.Sleep(2 * time.Second)

	return s
}

func (s *Subscriber) Stop() {
	for i, ps := range s.rps {
		if err := ps.Close(); err != nil {
			log.Err(err).
				Stringer("channel", channel(i)).
				Msg("failed to close subscriber")
		}
	}
	s.wg.Wait()
}

func (s *Subscriber) do(c channel, payload []byte) {

	log.Info().Stringer("channel", c).Msg("handling payload")

	var r trait.HasRoute

	switch c {
	case bots:
		r = &model.Result{}
	case news:
		r = &model.News{}
	case pages:
		r = &model.Page{}
	case sitemaps:
		r = &model.Sitemap{}
	case searches:
		r = &model.Search{}
	default:
		log.Warn().Msgf("unknown channel [%s]", c)
		return
	}

	if err := json.Unmarshal(payload, r); err != nil {
		return
	}

	switch c {
	case searches:
		s3.Put(r.(trait.HasContent).Content())
		s3.Put(r.(trait.HasScreenshot).Screenshot())
	case pages:
		s3.Put(r.(trait.HasScreenshot).Screenshot())
	default:
	}

	client.Put(os.Getenv("WEB_URL")+r.Route(), r)
}
