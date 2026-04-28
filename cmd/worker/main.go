package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/internal/worker"
	"github.com/nelsw/bytelyon/pkg/aws"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func main() {

	var ak, sk, u, l string
	flag.StringVar(&u, "u", "01KMXGBJJE2GMCA1A9EXDGF4AJ", "user id")
	flag.StringVar(&l, "l", "debug", "log level [trace, debug, info, warn, error]")
	flag.StringVar(&ak, "ak", "", "AWS Access Key ID")
	flag.StringVar(&sk, "sk", "", "AWS Secret Access Key")
	flag.Parse()

	logs.Init(l)
	logs.PrintWorkerBanner(map[string]any{
		"AWS Access Key": ak,
		"AWS Secret Key": sk,
		"Log Level":      l,
		"Process ID":     os.Getpid(),
		"User ID":        u,
	})

	aws.Init(ak, sk, "us-east-1")

	w := worker.New(ulid.MustParse(u))
	go w.Start()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("listening for quit signal (Ctrl+C)")
	<-quit
	fmt.Println()
	log.Info().Msg("quitting")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.Stop(ctx); err != nil {
		log.Err(err).Send()
	}

	<-ctx.Done()
	log.Info().Msg("exiting")
}
