package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	log.Logger = Make(zerolog.InfoLevel)
	
}

func Make(lvl zerolog.Level, out ...io.Writer) (logger zerolog.Logger) {

	if len(out) == 0 || out[0] == nil {
		out = append(out, defaultWriter())
	}

	logger = zerolog.New(out[0]).Level(lvl).With().Timestamp().Logger()

	if lvl == zerolog.TraceLevel {
		logger = logger.With().Caller().Logger()
	}

	return
}

func defaultWriter() io.Writer {
	return zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Kitchen,
		FieldsOrder: []string{
			"#", "id",
			"t", "type",
			"q", "query",
		},
		FormatLevel: func(a any) string {
			if a == nil || a == "<nil>" {
				a = "   "
			}
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "TRA":
				return "\033[0;36m" + l + "\033[0m"
			case "DEB":
				return "\033[0;35m" + l + "\033[0m"
			case "INF":
				return "\033[0;32m" + l + "\033[0m"
			case "WAR":
				return "\033[0;33m" + l + "\033[0m"
			case "ERR":
				return "\033[0;31m" + l + "\033[0m"
			case "FAT", "PAN":
				return "\033[41m" + "\033[0;37m" + l + "\033[0m"
			default:
				return ""
			}
		},
	}
}
