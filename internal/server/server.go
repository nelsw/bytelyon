package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/nelsw/bytelyon/internal/handler"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Server struct {
	port int
	*http.Server
}

func New(mode string, port int, db *gorm.DB) *Server {
	return &Server{
		port,
		&http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler.New(mode, db),
		},
	}
}

func (s *Server) Serve() {
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Int("port", s.port).Msg("Server failure")
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Err(err).Int("port", s.port).Msg("Server Shutdown")
	}
}
