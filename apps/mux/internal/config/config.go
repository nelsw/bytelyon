package config

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/nelsw/bytelyon/apps/mux/internal/logger"
	"github.com/nelsw/bytelyon/apps/mux/internal/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	var level, host, key string
	flag.StringVar(&level, "log", zerolog.TraceLevel.String(), "log level trace->disabled")
	flag.StringVar(&host, "host", "http://localhost", "web app (web) host name")
	flag.StringVar(&key, "key", "my-random-32-character-x-api-key", "web app (web) api key")
	flag.Parse()

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		panic(err)
	}

	log.Logger = logger.Make(lvl)
	service.WebHdr = http.Header{"x-api-key": []string{key}}
	service.WebUrl = fmt.Sprintf("%s/api", host)
}

func Print() {
	log.Log().Msgf("🦁")
	log.Log().Msg(`🦁  ByteLyon Mux (config)`)
	log.Log().Str("web.url", service.WebUrl).Msg(`🦁 `)
	log.Log().Str("web.key", service.WebHdr["x-api-key"][0]).Msg(`🦁 `)
	log.Log().Stringer("log.lvl", log.Logger.GetLevel()).Msg(`🦁 `)
	log.Log().Msgf("🦁\n")
}
