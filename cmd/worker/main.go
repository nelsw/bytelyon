package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/worker"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

// var pwc *playwright.Playwright
var uid ulid.ULID

func init() {

	//pwc = pw.Run()

	var u, l string
	flag.StringVar(&u, "u", ulid.Zero.String(), "user id")
	flag.StringVar(&l, "l", "debug", "log level [trace, debug, info, warn, error]")
	flag.Parse()

	logs.Init(l)
	uid = ulid.MustParse(u)
	m := map[string]any{
		"Log Level":  l,
		"Process ID": os.Getpid(),
		"User ID":    uid,
	}

	logs.Init(l)
	log.Info().Fields(m).Msg("starting")
	logs.PrintWorkerBanner(m)
}

func main() {
	pwc := pw.Run()
	// todo - try n instances
	w := worker.New(pwc, uid)
	go w.Start()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("listening for quit signal (Ctrl+C)")
	<-quit
	fmt.Println()

	log.Info().Msg("quitting")
	w.Stop()
	log.Info().Msg("exiting")
}
