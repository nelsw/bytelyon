package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/internal/job"
	"github.com/nelsw/bytelyon/apps/mux/internal/logger"
	"github.com/nelsw/bytelyon/apps/mux/internal/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	var logLvl, webHost string
	var broPort int
	flag.StringVar(&logLvl, "log", zerolog.TraceLevel.String(), "log level trace->disabled")
	flag.IntVar(&broPort, "bro", 8085, "bro app port")
	flag.StringVar(&webHost, "web", "http://localhost", "web app (web) host name")
	flag.Parse()

	lvl, err := zerolog.ParseLevel(logLvl)
	if err != nil {
		panic(err)
	}

	log.Logger = logger.Make(lvl)
	service.BroApiUrl = fmt.Sprintf("http://localhost:%d/bots", broPort)
	service.WebApiUrl = fmt.Sprintf("%s/api/bots", webHost)

	log.Log().Msgf("🦁")
	log.Log().Msg(`🦁  ByteLyon Mux (config)`)
	log.Log().Str("bro", service.BroApiUrl).Msg(`🦁 `)
	log.Log().Str("web", service.WebApiUrl).Msg(`🦁 `)
	log.Log().Stringer("log", log.Logger.GetLevel()).Msg(`🦁 `)
	log.Log().Msgf("🦁\n")
}

func main() {

	q := job.NewQueue()

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /bot", job.WebHandler(q))
	mux.HandleFunc("DELETE /bot/{id}", job.BroHandler(q))

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

	if err := server.Shutdown(context.Background()); err != nil {
		log.Err(err).Msg("failed to gracefully shutdown server")
	}

	t.Stop()
	q.Close()
	wg.Wait()

	fmt.Printf("\n👋\n")
}
