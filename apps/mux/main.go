package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	tx sync.Mutex
	wg sync.WaitGroup
	ip = make(map[int]bool)
	ch = make(chan *Bot)
	ctx = context.Background()
)

func main() {
	printBanner()
	poll()
	work()
	svr := listen()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	if err := svr.Shutdown(ctx); err != nil {
		log.Err(err).Msg("failed to gracefully shutdown server")
	}
	close(ch)
	wg.Wait()
	fmt.Printf("\n👋\n")
}

func listen() *http.Server {
	svr := &http.Server{Addr: ":" + os.Getenv("MUX_PORT")}

	http.HandleFunc("PUT /bot", func(w http.ResponseWriter, r *http.Request) {
		var bot Bot
		if err := json.NewDecoder(r.Body).Decode(&bot); err != nil {
			log.Warn().Err(err).Msg("Failed to decode bot")
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			ch <- &bot
			w.WriteHeader(http.StatusCreated)
		}
	})

	http.HandleFunc("DELETE /bot/{id}", func(w http.ResponseWriter, r *http.Request) {
		if botID, err := strconv.Atoi(r.PathValue("id")); err != nil {
			log.Warn().Err(err).Msgf("Invalid bot ID: %s", r.PathValue("id"))
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			tx.Lock()
			defer tx.Unlock()
			delete(ip, botID)
			w.WriteHeader(http.StatusNoContent)
		}
	})

	if err := svr.ListenAndServe(); err != nil {
		panic(err)
	}

	log.Info().Msg("listening...")
	return svr
}

func poll() {
	log.Info().Msg("polling...")
	for _, b := range Bots() {
		ch <- b
	}
	time.AfterFunc(5*time.Minute, poll)
}

func work() {
	ƒ := func(b *Bot) {
		tx.Lock()
		defer tx.Unlock()
		if _, exists := ip[b.ID]; !exists {
			ip[b.ID] = true
			b.Run()
		}
	}
	for i := 0; i < 3; i++ {
		wg.Go(func() {
			for b := range ch {
				ƒ(b)
			}
		})
	}
	log.Info().Msg("working...")
}
