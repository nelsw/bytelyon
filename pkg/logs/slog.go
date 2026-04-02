package logs

import (
	"log/slog"
	"os"

	slogzerolog "github.com/samber/slog-zerolog/v2"
)

func NewSlog() *slog.Logger {

	var sl slog.Level
	if os.Getenv("MODE") == "release" {
		sl = slog.LevelWarn
	} else if os.Getenv("MODE") == "debug" {
		sl = slog.LevelInfo
	} else {
		sl = slog.LevelDebug
	}

	return slog.New(slogzerolog.Option{
		Level:  sl,
		Logger: NewZerolog(),
	}.NewZerologHandler())
}
