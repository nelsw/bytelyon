package logs

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewZerolog() *zerolog.Logger {
	l := MakeZerolog()
	return &l
}

func MakeZerolog() zerolog.Logger {
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
		FieldsOrder: []string{
			"ƒ",
			"userId", "botId", "id",
			"type", "botType",
			"ready",
			"size",
			"table",
			"ip", "method", "authorization", "path", "query", "body",
		},
	})

	if os.Getenv("LOG_LEVEL") == "release" {
		return l.Level(zerolog.InfoLevel)
	}

	if os.Getenv("LOG_LEVEL") == "debug" {
		return l.Level(zerolog.DebugLevel)
	}

	return l.Level(zerolog.TraceLevel).With().Caller().Logger()
}
