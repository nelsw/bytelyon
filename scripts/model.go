package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Bot struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Query     string    `json:"query"`
	Headless  bool      `json:"headless"`
	LastRunAt time.Time `json:"last_run_at"`
	SitemapID int       `json:"sitemap_id"`
	SearchID  int       `json:"serp_id"`
	Blacklist []string  `json:"blacklist"`
}

func (b *Bot) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", b.ID).
		Str("q", b.Query).
		Any("t", b.Type)
}

// Run executes the prowl.py command with the bot's query and arguments
// todo - consider making the final call here as it's the safest place
// to determine if the command succeeded or failed.
func (b *Bot) Run(done func(b *Bot)) {

    defer done(b)

	args := []string{
		"-i", strconv.Itoa(b.ID),
		"-t", b.Type,
		"-q", b.Query,
		"-a", b.LastRunAt.Format(time.RFC3339),
		"--key", req.Header.Get("x-api-key"),
	}

	if bl := strings.TrimSpace(strings.Join(b.Blacklist, ",")); bl != "" {
		args = append(args, "-b", bl)
	}
	if b.Headless {
		args = append(args, "--headless")
	}
	if b.SitemapID > 0 {
		args = append(args, "-m", strconv.Itoa(b.SitemapID))
	}
	if b.SearchID > 0 {
		args = append(args, "-x", strconv.Itoa(b.SearchID))
	}

	if out, err := exec.Command("./prowl.py", args...).CombinedOutput(); err != nil {
		log.Err(err).Str("args", strings.Join(args, " ")).Msg("Command error")
	} else {
		log.Info().Msgf("Command output: %s", string(out))
	}
}

type Queue struct {
	sync.Mutex
	sync.WaitGroup
	wip map[int]bool
	less chan *Bot
	full chan *Bot
	tick *time.Ticker
}

func Start(
    headlessWorkers int,
    headfullWorkers int,
    pollingInterval int,
) *Queue {
	q := &Queue{
	    wip: make(map[int]bool),
	    less: make(chan *Bot),
	    full: make(chan *Bot),
	    tick: time.NewTicker(time.Second * time.Duration(pollingInterval)),
	}
	q.work(q.full, headfullWorkers)
	q.work(q.less, headfullWorkers)
	go q.Poll()
	return q
}

func (q *Queue) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("name", "queue").
		Int("wip", len(q.wip)).
		Int("less", len(q.less)).
		Int("full", len(q.full))
}

func (q *Queue) Quit() {
	q.tick.Stop()
	close(q.less)
	close(q.full)
	q.Wait()
}

func (q *Queue) check(b *Bot) (cb func(b *Bot), ok bool) {
	q.Lock()
	defer q.Unlock()
	if _, ok = q.wip[b.ID]; !ok {
		q.wip[b.ID] = true
	}
	return func(b *Bot) {
    	q.Lock()
    	defer q.Unlock()
    	delete(q.wip, b.ID)
	}, ok
}

func (q *Queue) Poll() {
    for range q.tick.C {
       	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Err(err).Msg("Failed to get bots")
	}
	defer func() {
		_ = res.Body.Close()
	}()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		log.Err(err).Msg("Failed to read response body")
	}

	var bots []Bot
	if err = json.Unmarshal(body, &bots); err != nil {
		log.Err(err).Bytes("body", body).Msg("Failed to unmarshal bots")
	}

	for _, b := range bots {
		if b.Headless {
			q.less <- &b
		} else {
			q.full <- &b
		}
	}
	log.Info().EmbedObject(q).Send()
    }
}

func (q *Queue) work(c chan *Bot, count int) func() {
    return func () {
        for range count {
            q.Go(func() {
                for b := range c {
        			if cb, ok := q.check(b); ok {
        			    b.Run(cb)
        			}
        		}
            })
        }
    }
}
