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

	"github.com/nelsw/bytelyon/config"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/manager"
	"github.com/nelsw/bytelyon/internal/router"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.New()

	logger.Init(cfg.Mode)

	DB := db.New(cfg.Mode)

	r := router.New(cfg.Mode, DB)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: r.Handler(),
	}

	mgr := manager.New(DB)

	go func() {
		mgr.Start()
		// service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Msgf("listen: %s\n", err)
		}
	}()

	// Wait for the interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down Server ...")

	mgr.Stop()
	for !mgr.Done() {
		time.Sleep(time.Second)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Err(err).Msg("Server Shutdown")
	}
	<-ctx.Done()
	log.Info().Msg("Server exiting")
}
