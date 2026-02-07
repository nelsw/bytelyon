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
	"github.com/nelsw/bytelyon/internal/handler"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.New()

	gdb := db.New(cfg.Mode)
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: handler.New(cfg.Mode, gdb),
	}

	go func() {
		if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Int("port", cfg.Port).Msg("Server failure")
		}
	}()
	log.Info().Int("port", cfg.Port).Msg("Server listening")

	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println() // Print a newline after the signal is received to escape cmd

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Int("port", cfg.Port).Msg("Server stopping")
	if err := svr.Shutdown(ctx); err != nil {
		log.Err(err).Int("port", cfg.Port).Msg("Server Shutdown")
	}

	<-ctx.Done()
	log.Info().Int("port", cfg.Port).Msg("Server exiting")
}
