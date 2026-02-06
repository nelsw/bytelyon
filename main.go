package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/config"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/server"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.New()

	DB := db.New(cfg.Mode)
	//mgr := manager.New(DB)
	srv := server.New(cfg.Mode, cfg.Port, DB)

	//go mgr.Start()
	//log.Info().Int("port", cfg.Port).Msg("Manager started")

	go srv.Serve()
	log.Info().Int("port", cfg.Port).Msg("Server listening")

	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println() // Print a newline after the signal is received to escape cmd
	//
	//log.Info().Int("port", cfg.Port).Msg("Manager stopping")
	//for !mgr.Stop() {
	//	time.Sleep(time.Second)
	//}
	//log.Info().Int("port", cfg.Port).Msg("Manager stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Int("port", cfg.Port).Msg("Server stopping")
	srv.Shutdown(ctx)

	<-ctx.Done()
	log.Info().Int("port", cfg.Port).Msg("Server exiting")
}
