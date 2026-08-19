package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

var url string
var key string

func init() {
	flag.StringVar(&url, "url", "http://localhost:80", "URL to poll")
	flag.StringVar(&key, "key", "my-random-32-character-x-api-key", "API key for authentication")
	flag.Parse()
}

func main() {
	log.Println("starting...")

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
			log.Printf("\nstopping...\n")
			tick.Stop()
			close(full)
			close(less)
			wg.Wait()
			log.Println("👋")
			return
		}
	}
}

func poll(full, less chan *Bot) {
	log.Println("polling...")

	req, _ := http.NewRequest("GET", url+"/api/bots", nil)
	req.Header.Set("x-api-key", key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to get bots: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		log.Printf("Failed to read response body: %v", err)
	}

	var bots []Bot
	fmt.Println("Bots:", string(body))
	if err = json.Unmarshal(body, &bots); err != nil {
		log.Printf("Failed to unmarshal bots: %v", err)
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

				combined(b)

				//data, _ := json.Marshal(b)
				//
				//cmd := exec.Command("./main.py", string(data))
				//
				//stdout, err := cmd.StdoutPipe()
				//if err != nil {
				//	log.Printf("Failed to create stdout pipe: %v\n", err)
				//	return
				//}
				//
				//if err = cmd.Start(); err != nil {
				//	log.Printf("Failed to start command: %v\n", err)
				//	return
				//}
				//
				//scanner := bufio.NewScanner(stdout)
				//for scanner.Scan() {
				//	fmt.Printf("%s\n", scanner.Text())
				//}
				//
				//if err = scanner.Err(); err != nil {
				//	log.Printf("scanner err: %v\n", err)
				//} else if err = cmd.Wait(); err != nil {
				//	log.Printf("Command finished with error: %v\n", err)
				//}
			}
		}
	}

	wg.Go(ƒ(full))
	for range 3 {
		wg.Go(ƒ(less))
	}
	log.Println("working...")
}

func combined(b *Bot) {
	data, _ := json.Marshal(b)
	cmd := exec.Command("./main.py", string(data))
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command finished with error: %v\n", err)
	} else {
		fmt.Println(string(out))
	}
	fmt.Println(string(out))

}
