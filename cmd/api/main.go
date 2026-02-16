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

	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/router"
	"github.com/rs/zerolog/log"
)

func init() {
	config.Init("API Server")
	logger.Init()
	db.Init()
}

func main() {

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port()),
		Handler: router.New().Handler(),
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Int("port", config.Port()).Msg("Server stopping")
	if err := svr.Shutdown(ctx); err != nil {
		log.Err(err).Int("port", config.Port()).Msg("Server Shutdown")
	}

	<-ctx.Done()
	log.Info().Int("port", config.Port()).Msg("Server exiting")
}
