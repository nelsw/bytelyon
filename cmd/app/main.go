package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon/internal/client/prowl"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/manager"
	"github.com/nelsw/bytelyon/internal/router"
	"github.com/rs/zerolog/log"
)

func init() {
	config.Init()
	logger.Init()
	prowl.Init()
}

func main() {

	mgr := manager.New()
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port()),
		Handler: router.New().Handler(),
	}

	go mgr.Start()
	go func() {
		if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Int("port", config.Port()).Msg("Server failure")
		}
	}()
	log.Info().Int("port", config.Port()).Msg("Server listening")

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

	log.Info().Int("port", config.Port()).Msg("Server stopping")
	if err := svr.Shutdown(ctx); err != nil {
		log.Err(err).Int("port", config.Port()).Msg("Server Shutdown")
	}

	<-ctx.Done()
	log.Info().Int("port", config.Port()).Msg("Server exiting")
}
