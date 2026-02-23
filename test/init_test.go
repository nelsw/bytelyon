package test

import (
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/logger"
)

func init() {
	config.Init()
	logger.Init()
	db.Init()
}
