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

func main() {

	var tkn, store string
	flag.StringVar(&tkn, "token", "", "Shopify API token")
	flag.StringVar(&store, "store", "", "Shopify store ID")
	flag.Parse()

	logs.Init()
	w := snooper.New("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", "e61745-7d")
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
