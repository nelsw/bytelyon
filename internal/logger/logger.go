package logger

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(mode string) {
	log.Logger = Make(mode)
}

func New(mode string) *zerolog.Logger {
	l := Make(mode)
	return &l
}

func Make(mode string) zerolog.Logger {

	l := log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatLevel: func(a any) string {
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "TRA":
				return Cyan + l + Default
			case "DEB":
				return Purple + l + Default
			case "INF":
				return Green + l + Default
			case "WAR":
				return Yellow + l + Default
			case "ERR":
				return Red + l + Default
			case "FAT", "PAN":
				return RedBackground + White + l + Default
			default:
				return Default + l + Default
			}
		},
	})

	switch mode {
	case gin.ReleaseMode:
		return l.Level(zerolog.InfoLevel)
	case gin.DebugMode:
		return l.Level(zerolog.DebugLevel)
	}

	return l.Level(zerolog.TraceLevel).With().Caller().Logger()
}
