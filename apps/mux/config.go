package main

import (
	"io"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	level, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = zerolog.DebugLevel
	}

	initLogger(level, nil)
}

func initLogger(level zerolog.Level, out io.Writer) {
	if out == nil {
		out = zerolog.ConsoleWriter{
			Out:         os.Stdout,
			FieldsOrder: []string{},
			FormatLevel: func(a any) string {
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
					return "\033[0m" + l + "\033[0m"
				}
			},
		}
	}
	if log.Logger = zerolog.New(out).Level(level); level == zerolog.TraceLevel {
		log.Logger = log.Logger.With().Caller().Logger()
	}
}

func printBanner() {
	log.Log().Msgf("\n🦁")
	log.Log().Msg(`🦁  ByteLyon Muxer`)
	log.Log().Str("bro.url", os.Getenv("BRO_URL")).Msg(`🦁 `)
	log.Log().Str("mux.url", os.Getenv("MUX_URL")).Msg(`🦁 `)
	log.Log().Str("web.url", os.Getenv("WEB_URL")).Msg(`🦁 `)
	log.Log().Str("web.key", "..."+os.Getenv("WEB_KEY")[26:]).Msg(`🦁 `)
	log.Log().Stringer("log lvl", log.Logger.GetLevel()).Msg(`🦁 `)
	log.Log().Msgf("🦁\n")
}
