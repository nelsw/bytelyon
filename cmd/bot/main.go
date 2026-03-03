package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/internal/client/prowl"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/manager"
	"github.com/rs/zerolog/log"
)

func init() {
	config.Init()
	logger.Init()
	prowl.Init()
}

func main() {

	mgr := manager.New()

	go mgr.Start()

	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := mgr.Stop(ctx); err != nil {
		log.Err(err).Msg("Manager stop failure")
	}

	<-ctx.Done()
	log.Info().Int("port", config.Port()).Msg("Server exiting")
}
