package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/internal/client"
	"github.com/nelsw/bytelyon/apps/mux/internal/config"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/nelsw/bytelyon/apps/mux/internal/queue"
	"github.com/rs/zerolog/log"
)

func main() {

	config.Print()

	q := queue.New()

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /bot", func(w http.ResponseWriter, r *http.Request) {
		var bot model.Bot
		if err := json.NewDecoder(r.Body).Decode(&bot); err != nil {
			log.Warn().Err(err).Msg("Failed to decode bot")
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			q.Send(&bot)
			w.WriteHeader(http.StatusCreated)
		}
	})

	server := &http.Server{
		Addr:         ":3000",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	log.Info().Msg("listening...")

	t := time.NewTicker(5 * time.Minute)
	go func() {
		log.Info().Msg("polling...")
		q.Send(client.Get[model.Bots]()...)
		for range t.C {
			q.Send(client.Get[model.Bots]()...)
		}
	}()

	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			for b := range q.Chan() {
				if q.Put(b.ID) {
					client.Post("http://localhost:8085/bots", b)
				}
			}
		})
	}
	log.Info().Msg("working...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	if err := server.Shutdown(context.Background()); err != nil {
		log.Err(err).Msg("failed to gracefully shutdown server")
	}

	t.Stop()
	q.Close()
	wg.Wait()

	fmt.Printf("\n👋\n")
}
