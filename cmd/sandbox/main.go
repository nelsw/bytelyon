package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func main() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("listening for quit signal (Ctrl+C)")
	<-quit
	fmt.Println()

	log.Info().Msg("quitting")
	log.Info().Msg("exiting")
}
