package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nelsw/bytelyon/internal/snooper"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/rs/zerolog/log"
)

var tkn, store, l string

func init() {
	flag.StringVar(&tkn, "token", "", "Shopify API token")
	flag.StringVar(&store, "store", "", "Shopify store ID")
	flag.StringVar(&l, "l", "debug", "log level [trace, debug, info, warn, error]")
	flag.Parse()

	logs.Init(l)
	logs.PrintBanner()
}

func main() {
	w := snooper.New(tkn, store)
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
