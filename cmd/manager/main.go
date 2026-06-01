package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/internal/sys"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/rs/zerolog/log"
)

func main() {

	var l string
	flag.StringVar(&l, "l", "debug", "log level [trace, debug, info, warn, error]")
	flag.Parse()

	m := map[string]any{
		"Log Level":  l,
		"Process ID": os.Getpid(),
	}

	logs.Init(l)
	log.Info().Fields(m).Msg("starting")
	logs.PrintWorkerBanner(m)

	mgr, err := sys.NewManager()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create manager")
	}
	go mgr.Start()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("listening for quit signal (Ctrl+C)")
	<-quit
	fmt.Println()
	log.Info().Msg("quitting")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = mgr.Stop(ctx); err != nil {
		log.Err(err).Msg("failed to stop manager")
	}

	<-ctx.Done()
	log.Info().Msg("exiting")
}
