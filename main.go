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

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/rs/zerolog/log"
)

func init() {

}

func main() {

	cfg := config.New()

	gin.SetMode(cfg.Mode)
	gin.ForceConsoleColor()

	//DB := db.New(cfg.Mode)

	log.Logger = logger.Make(cfg.Mode)

	r := gin.New()

	r.LoadHTMLGlob("web/templates/*")
	//r.Static("/assets", "./assets")
	r.GET("/", func(c *gin.Context) {
		var data any
		c.HTML(http.StatusOK, "page.gohtml", data)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: r.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Msgf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Err(err).Msg("Server Shutdown")
	}
	<-ctx.Done()
	log.Info().Msg("Server exiting")
}
