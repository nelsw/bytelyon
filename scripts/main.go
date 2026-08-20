package main

import (
	"bufio"
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
	poller   *time.Ticker
	working  map[int]bool
	headless chan *Bot
	headfull chan *Bot
	lessLen  int
	fullLen  int
	pollDur  int
	key      string
	api      string
}

func (q *Queue) Start() *Queue {

	q.poller = time.NewTicker(time.Duration(q.pollDur) * time.Second)

	ƒ := func(count int, c chan *Bot) {
		for range count {
			q.Go(func() {
				for b := range c {
					if ran, wip := q.check(b); wip {
						_ = b.Run(q.key)
						ran(b)
					}
				}
			})
		}
	}

	ƒ(q.fullLen, q.headfull)
	ƒ(q.lessLen, q.headless)

	return q
}

func (q *Queue) Stop() {
	q.poller.Stop()
	close(q.headless)
	close(q.headfull)
	q.Wait()
}

func (q *Queue) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("🚜", len(q.working)).
		Int("🐺", len(q.headless)).
		Int("🦄", len(q.headfull))
}

func (q *Queue) check(b *Bot) (done func(b *Bot), ok bool) {
	q.Lock()
	defer q.Unlock()
	if q.working[b.ID] {
		return
	}
	q.working[b.ID] = true
	return func(b *Bot) {
		q.Lock()
		defer q.Unlock()
		delete(q.working, b.ID)
		log.Debug().EmbedObject(b).Msgf("done")
	}, true
}

func (q *Queue) Poll() {

	req, _ := http.NewRequest("GET", q.api+"/bots", nil)
	req.Header.Set("x-api-key", q.key)

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
			q.headless <- &b
		} else {
			q.headfull <- &b
		}
	}
	log.Log().EmbedObject(q).Send()
}

type BotType string

const (
	NewsBot    BotType = "news"
	SearchBot  BotType = "search"
	SitemapBot BotType = "sitemap"
)

func (t *BotType) String() string { return string(*t) }
func (t *BotType) UnmarshalJSON(payload []byte) error {
	if text := string(payload); text == `"news"` || text == `"search"` || text == `"sitemap"` {
		*t = BotType(strings.ReplaceAll(text, `"`, ""))
		return nil
	}
	return fmt.Errorf("unknown bot type: %s", payload)
}

type Bot struct {
	ID        int       `json:"id"`
	Type      BotType   `json:"type"`
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

func (b *Bot) Run(key string) error {

	args := []string{
		"-i", strconv.Itoa(b.ID),
		"-q", fmt.Sprintf(`'%s'`, b.Query),
		"-k", key,
	}
	if b.Headless {
		args = append(args, "--headless")
	}

	switch b.Type {
	case NewsBot:
		args = append(args, "-a", b.LastRunAt.Format(time.RFC3339))
	case SitemapBot:
		args = append(args, "-x", strconv.Itoa(b.SitemapID))
	case SearchBot:
		args = append(args, "-x", strconv.Itoa(b.SearchID))
		if bl := strings.TrimSpace(strings.Join(b.Blacklist, ",")); bl != "" {
			args = append(args, "-b", bl)
		}
	}

	name := fmt.Sprintf("./%s_bot.py", b.Type)

	log.Debug().Str("cmd", name+" "+strings.Join(args, " ")).Send()

	cmd := exec.Command(name, args...)

	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	} else if err = cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		log.Log().Msg(scanner.Text())
	}
	return cmd.Wait()
}

func main() {

	q := &Queue{
		working:  make(map[int]bool),
		headless: make(chan *Bot),
		headfull: make(chan *Bot),
	}

	flag.StringVar(&q.key, "key", "my-random-32-character-x-api-key", "API Auth Key")
	flag.IntVar(&q.pollDur, "poll", 5, "Polling Interval in seconds")
	flag.IntVar(&q.fullLen, "full", 1, "Number of headfull workers")
	flag.IntVar(&q.lessLen, "less", 3, "Number of headless workers")
	flag.Parse()

	if len(q.key) == 36 {
		q.api = "https://bytelyon.com/api"
	} else {
		q.api = "http://localhost:80/api"
	}

	log.Logger = makeLogger()

	log.Log().Msg(`🦁 `)
	log.Log().Msg(`🦁  ByteLyon Bot Runner`)
	log.Log().Str("api", q.api).Msg(`🦁 `)
	log.Log().Str("key", q.key).Msg(`🦁 `)
	log.Log().Msg(`🦁 `)

	q.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-q.poller.C:
			q.Poll()
		case <-quit:
			fmt.Println() // newline for ^C buffer entry
			q.Stop()
			log.Log().Msg("👋")
			return
		}
	}
}

func makeLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FieldsOrder: []string{
			"time", "level", "msg", // general fields
			"🚜", "🦄", "🐺", // embedded q
			"#", "t", "q", // embedded bot
			"cmd", // always last
		},
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
