package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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

var req *http.Request

func init() {

    var profile string
	flag.StringVar(&profile, "profile", "", "local (default), testing, production")
	flag.Parse()

	file := "../.env"
	if profile != "" {
		file = "../.env." + profile
	}

	if err := godotenv.Load(file); err != nil {
		panic(err)
	}

	req, _ = http.NewRequest("GET", os.Getenv("API_URL")+"/api/bots", nil)
	req.Header.Set("x-api-key", os.Getenv("API_KEY"))

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
}

func main() {
	log.Info().Msg("starting...")

	full, less := make(chan *Bot), make(chan *Bot)

	var wg sync.WaitGroup
	work(&wg, full, less)

	tick := time.NewTicker(5 * time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-tick.C:
			poll(full, less)
		case <-quit:
		    fmt.Println()
			log.Info().Msg("stopping...")
			tick.Stop()
			close(full)
			close(less)
			wg.Wait()
			log.Info().Msg("👋")
			return
		}
	}
}

func poll(full, less chan *Bot) {
	log.Info().Msg("polling...")

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
		log.Err(err).Msg("Failed to unmarshal bots")
	}

	for _, b := range bots {
		if b.Headless {
			less <- &b
		} else {
			full <- &b
		}
	}
}

func work(wg *sync.WaitGroup, full, less chan *Bot) {

	ƒ := func(c chan *Bot) func() {
		return func() {
			for b := range c {
    			args := []string{
                    strconv.Itoa(b.ID),
                    b.Type,
                    b.Query,
                    strings.Join(b.Blacklist, ","),
                    b.LastRunAt.Format(time.RFC3339),
                    strconv.FormatBool(b.Headless),
                    strconv.Itoa(b.SitemapID),
                    strconv.Itoa(b.SearchID),
                }

                if out, err := exec.Command("./main.py", args...).CombinedOutput(); err != nil {
                   	log.Printf("Command error: %v\n", err)
                } else {
                    log.Printf("Command output: %s\n", string(out))
                }
			}
		}
	}

	wg.Go(ƒ(full))
	for range 3 {
		wg.Go(ƒ(less))
	}
	log.Info().Msg("working...")
}
