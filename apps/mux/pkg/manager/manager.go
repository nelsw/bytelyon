package manager

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/pkg/bro"
	"github.com/nelsw/bytelyon/apps/mux/pkg/model"
	"github.com/nelsw/bytelyon/apps/mux/pkg/s3"
	"github.com/nelsw/bytelyon/apps/mux/pkg/web"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	sync.RWMutex
	bro *bro.Client
	web *web.Client
	que map[int]*model.Bot
}

func New(broPort int, webHost, webKey string) *Manager {
	return &Manager{
		bro: bro.New(broPort),
		web: web.New(webHost, webKey),
		que: make(map[int]*model.Bot),
	}
}

func (m *Manager) Work() {

	bots, err := m.web.GetBots()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get bots")
		return
	} else if len(bots) == 0 {
		log.Info().Msg("no bots to work on")
		return
	}

	log.Info().Msgf("starting work [%d]", len(bots))

	m.Lock()
	defer m.Unlock()

	for _, bot := range bots {
		if _, ok := m.que[bot.ID]; !ok {
			m.que[bot.ID] = bot
			go func() {
				if err = m.bro.Put(bot); err != nil {
					m.pop(bot.ID)
					m.web.PutBot(bot, err)
				}
				log.Err(err).
					EmbedObject(bot).
					Msg("bot work requested")
			}()
		}
	}
}

func (m *Manager) Quit() {
	busy := func() bool {
		m.Lock()
		defer m.Unlock()
		return len(m.que) > 0
	}
	for busy() {
		time.Sleep(time.Second * 3)
	}
}

func (m *Manager) pop(id int) *model.Bot {
	m.Lock()
	defer func() {
		m.Unlock()
		delete(m.que, id)
	}()
	if _, exists := m.que[id]; exists {
		return m.que[id]
	}
	return nil
}

func (m *Manager) push(b *model.Bot) bool {
	m.Lock()
	defer m.Unlock()
	if _, exists := m.que[b.ID]; exists {
		return false
	}
	m.que[b.ID] = b
	return true
}

func (m *Manager) BroHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Warn().Msgf("Method %s not allowed", r.Method)
		}

		botID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		bot := m.pop(botID)

		var body []byte
		if body, err = io.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.PathValue("id") {
		case "news":
			var arr []any
			if err = json.Unmarshal(body, &arr); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				m.web.PutArticles(bot, arr)
			}
		case "search":
			var serp model.Serp
			var b []byte
			for _, v := range serp.Pages() {
				// tdodo set key
				s3.Save(v.Key, b)
				m.web.PutPage(bot, v)
			}
			//s3.Save(serp.SrcKey, serp.Src)
			//s3.Save(serp.ImgKey, serp.Img)
			//m.web.PutSerp(bot, serp)
		case "sitemap":
			var sitemap model.Sitemap
			var b []byte
			for _, v := range sitemap.Pages {
				s3.Save(v.Key, b)
				m.web.PutPage(bot, v)
			}
			m.web.PutSitemap(bot, sitemap)
		default:
			// todo - page
		}

	}
}

func (m *Manager) WebHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var bot model.Bot
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Warn().Msgf("Method %s not allowed", r.Method)
		} else if err := json.NewDecoder(r.Body).Decode(&bot); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if !m.push(&bot) {
			log.Warn().Msgf("Bot %d already exists", bot.ID)
			w.WriteHeader(http.StatusTooEarly)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
	}
}
