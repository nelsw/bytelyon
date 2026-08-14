package config

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

func MakeLogger(lvl zerolog.Level, out ...io.Writer) (logger zerolog.Logger) {

	if len(out) == 0 {
		out = append(out, defaultWriter())
	}

	logger = zerolog.New(out[0]).Level(lvl)
	
	if lvl == zerolog.TraceLevel {
		logger = logger.With().Caller().Logger()
	}
	
	return
}

func defaultWriter() io.Writer {
	return zerolog.ConsoleWriter{
		Out:         os.Stdout,
		FieldsOrder: []string{},
		FormatLevel: func(a any) string {
			switch fmt.Sprint(a) {
			case "trace":
				return "\033[0;36mTRA\033[0m"
			case "debug":
				return "\033[0;35mDEB\033[0m"
			case "info":
				return "\033[0;32mINF\033[0m"
			case "warn":
				return "\033[0;33mWAR\033[0m"
			case "error":
				return "\033[0;31mERR\033[0m"
			case "fatal":
				return "\033[41m\033[0;37mFAT\033[0m"
			case "panic":
				return "\033[41m\033[0;37mPAN\033[0m"
			default:
				return "\033[0m"
			}
		},
	}
}
