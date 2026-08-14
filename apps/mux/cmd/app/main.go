package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/internal/config"
	"github.com/nelsw/bytelyon/apps/mux/internal/job"
	"github.com/nelsw/bytelyon/apps/mux/internal/service"
	"github.com/rs/zerolog/log"
)

func main() {

	config.Print()

	q := job.NewQueue()

	svr := &http.Server{Addr: ":" + os.Getenv("MUX_PORT")}
	http.HandleFunc("PUT /bot", job.WebHandler(q))
	http.HandleFunc("DELETE /bot/{id}", job.BroHandler(q))
	if err := svr.ListenAndServe(); err != nil {
		panic(err)
	}
	log.Info().Msg("listening...")

	t := time.NewTicker(5 * time.Minute)
	go func() {
		for range t.C {
			log.Info().Msg("polling...")
			q.Send(service.GetWebBots()...)
		}
	}()

	var	wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			for b := range q.Chan() {
				if q.Put(b.ID) {
					service.PostBroBot(b)
				}
			}
		})
	}
	log.Info().Msg("working...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	if err := svr.Shutdown(context.Background()); err != nil {
		log.Err(err).Msg("failed to gracefully shutdown server")
	}

	t.Stop()
	q.Close()
	wg.Wait()

	fmt.Printf("\n👋\n")
}
