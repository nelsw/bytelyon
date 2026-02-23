package logger

import (
	"log/slog"

	"github.com/nelsw/bytelyon/internal/config"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

func NewSlog() *slog.Logger {

	var sl slog.Level
	if config.IsReleaseMode() {
		sl = slog.LevelError
	} else if config.IsDebugMode() {
		sl = slog.LevelInfo
	} else {
		sl = slog.LevelDebug
	}

	return slog.New(slogzerolog.Option{
		Level:  sl,
		Logger: NewZerolog(),
	}.NewZerologHandler())
}
