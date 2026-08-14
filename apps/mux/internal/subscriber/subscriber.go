package subscriber

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/internal/db"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/nelsw/bytelyon/apps/mux/internal/service"
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

	log.Info().
		Stringer("channel", c).
		Bytes("payload", payload).
		Msg("handling message")

	switch c {
	case bots:
		s.doBot(payload)
	case news:
		s.doNews(payload)
	case pages:
		s.doPages(payload)
	case sitemaps:
		s.doSitemaps(payload)
	case searches:
		s.doSearches(payload)
	default:
		log.Warn().Msgf("unknown channel [%s]", c)
	}
}

func (s *Subscriber) doBot(payload []byte) {

	var b model.Bot
	if err := json.Unmarshal(payload, &b); err != nil {
		log.Err(err).
			Bytes("payload", payload).
			Msg("failed to unmarshal bot")
		return
	}

	service.PutBot(&b, map[string]any{
		"result": "ok",
	})
}

func (s *Subscriber) doNews(payload []byte) {

}

func (s *Subscriber) doPages(payload []byte) {

}


func (s *Subscriber) doSearches(payload []byte) {

}

func (s *Subscriber) doSitemaps(payload []byte) {

}



