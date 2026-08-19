package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"encoding/json"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Queue struct {
	sync.Mutex
	sync.WaitGroup
	*time.Ticker
	wip  map[int]bool
	less chan *Bot
	full chan *Bot
}

func Start(
	headlessWorkers int,
	headfullWorkers int,
	pollingInterval int,
) *Queue {
	q := &Queue{
		wip:    make(map[int]bool),
		less:   make(chan *Bot),
		full:   make(chan *Bot),
		Ticker: time.NewTicker(time.Second * time.Duration(pollingInterval)),
	}

	ƒ := func(count int, c chan *Bot) {
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

	ƒ(headfullWorkers, q.full)
	ƒ(headlessWorkers, q.less)
	
	return q
}

func (q *Queue) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("wip", len(q.wip)).
		Int("less", len(q.less)).
		Int("full", len(q.full))
}

func (q *Queue) check(b *Bot) (func(b *Bot), bool) {
	q.Lock()
	defer q.Unlock()
	if q.wip[b.ID] {
		return nil, false
	}
	q.wip[b.ID] = true
	return func(b *Bot) {
		q.Lock()
		defer q.Unlock()
		delete(q.wip, b.ID)
	}, true
}

func (q *Queue) Poll() {
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

	log.Info().EmbedObject(b).Msg("running...")

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

var req *http.Request

func main() {

	var key string
	var poll, full, less int
	flag.StringVar(&key, "key", "my-random-32-character-x-api-key", "API Auth Key")
	flag.IntVar(&poll, "poll", 5, "Polling Interval in seconds")
	flag.IntVar(&full, "full", 1, "Number of full workers")
	flag.IntVar(&less, "less", 3, "Number of less workers")
	flag.Parse()

	if len(key) == 36 {
		req, _ = http.NewRequest("GET", "https://bytelyon.com/api/bots", nil)
	} else {
		req, _ = http.NewRequest("GET", "http://localhost:80/api/bots", nil)
	}
	req.Header.Set("x-api-key", key)

	log.Logger = makeLogger()

	log.Log().Msg(`🦁 `)
	log.Log().Msg(`🦁  ByteLyon Bot Runner`)
	log.Log().Str("api", req.URL.Host).Msg(`🦁 `)
	log.Log().Msg(`🦁 `)

	q := Start(less, full, poll)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-q.C:
			q.Poll()
		case <-quit:
			fmt.Println() // newline for ^C buffer entry
			q.Stop()
			close(q.less)
			close(q.full)
			q.Wait()
			log.Log().Msg("👋")
			return
		}
	}
}

func makeLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatLevel: func(a any) string {
			if a == nil || a == "<nil>" {
				a = "   "
			}
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "TRA":
				return "\033[0;36m" + l + "\033[0m"
			case "DEB":
				return "\033[0;35m" + l + "\033[0m"
			case "INF":
				return "\033[0;32m" + l + "\033[0m"
			case "WAR":
				return "\033[0;33m" + l + "\033[0m"
			case "ERR":
				return "\033[0;31m" + l + "\033[0m"
			case "FAT", "PAN":
				return "\033[41m" + "\033[0;37m" + l + "\033[0m"
			default:
				return ""
			}
		},
	}).With().Timestamp().Logger()
}
