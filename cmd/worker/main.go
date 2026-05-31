package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/user"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func main() {

	var u, l string
	flag.StringVar(&u, "u", ulid.Zero.String(), "user id")
	flag.StringVar(&l, "l", "debug", "log level [trace, debug, info, warn, error]")
	flag.Parse()

	m := map[string]any{
		"Log Level":  l,
		"Process ID": os.Getpid(),
		"User ID":    u,
	}

	logs.Init(l)
	log.Info().Fields(m).Msg("starting")
	logs.PrintWorkerBanner(m)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("listening for quit signal (Ctrl+C)")

	pwc := pw.Run()
	var stop bool
	go func() {
		for !stop {
			log.Info().Msg("running")
			user.Run(pwc, id.ParseULID(u))
			logs.PrintNyanCat()
			if !stop {
				time.Sleep(time.Second * 10)
			}
		}
		if err := pwc.Stop(); err != nil {
			log.Err(err).Send()
		}
	}()
	<-quit
	fmt.Println()

	log.Info().Msg("quitting")
	stop = true

	fmt.Println()
	log.Info().Msg("exiting")
}
