package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/config"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/manager"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.New()

	DB := db.New(cfg.Mode)
	mgr := manager.New(DB)

	go mgr.Start()
	log.Info().Int("port", cfg.Port).Msg("Manager started")

	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println() // Print a newline after the signal is received to escape cmd

	log.Info().Msg("Manager stopping")
	for !mgr.Stop() {
		time.Sleep(time.Second)
	}
	log.Info().Msg("Manager stopped")
}
