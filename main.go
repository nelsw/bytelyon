package main

import (
	"bufio"
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

var url string

func init() {
	flag.StringVar(&url, "url", "http://localhost:80", "URL to poll")
}

func main() {
	log.Println("starting...")

	full, less := make(chan []byte), make(chan []byte)

	var wg sync.WaitGroup
	work(&wg, full, less)

	tick := time.NewTicker(15 * time.Second)

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

func poll(full, less chan []byte) {
	log.Println("polling...")

	res, err := http.Get(url + "/api/bots")
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

	var bots []struct {
		ID        int       `json:"id"`
		Type      string    `json:"type"`
		Query     string    `json:"query"`
		Headless  bool      `json:"headless"`
		LastRunAt time.Time `json:"last_run_at"`
		SitemapID int       `json:"sitemap_id"`
		SearchID  int       `json:"serp_id"`
		Blacklist []string  `json:"blacklist"`
	}
	fmt.Println("Bots:", string(body))
	if err = json.Unmarshal(body, &bots); err != nil {
		log.Printf("Failed to unmarshal bots: %v", err)
	}

	for _, b := range bots {
		if b.Headless {
			less <- body
		} else {
			full <- body
		}
	}
}

func work(wg *sync.WaitGroup, full, less chan []byte) {

	ƒ := func(c chan []byte) func() {
		return func() {
			for b := range c {

				cmd := exec.Command("./main.py", string(b))

				stdout, err := cmd.StdoutPipe()
				if err != nil {
					log.Printf("Failed to create stdout pipe: %v\n", err)
					return
				}

				if err = cmd.Start(); err != nil {
					log.Printf("Failed to start command: %v\n", err)
					return
				}

				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					fmt.Printf("%s\n", scanner.Text())
				}

				if err = scanner.Err(); err != nil {
					log.Printf("scanner err: %v\n", err)
				} else if err = cmd.Wait(); err != nil {
					log.Printf("Command finished with error: %v\n", err)
				}
			}
		}
	}

	wg.Go(ƒ(full))
	for range 3 {
		wg.Go(ƒ(less))
	}
	log.Println("working...")
}
