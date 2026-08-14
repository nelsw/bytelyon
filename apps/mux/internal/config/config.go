package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/nelsw/bytelyon/apps/mux/internal/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {

	var level, host, key, bucket string

	flag.StringVar(&level, "log", "trace", "log level trace->disabled")
	flag.StringVar(&host, "host", "http://localhost", "web app (web) host name")
	flag.StringVar(&key, "key", "my-random-32-character-x-api-key", "web app (web) api key")
	flag.StringVar(&bucket, "s3", "bytelyon-private", "s3 bucket name")
	flag.Parse()

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.TraceLevel
	}

	log.Logger = logger.Make(lvl)
	_ = os.Setenv("WEB_KEY", key)
	_ = os.Setenv("WEB_URL", fmt.Sprintf("%s/api", host))
	_ = os.Setenv("S3_BUCKET", bucket)
}

func Print() {
	log.Log().Msgf("🦁")
	log.Log().Msg(`🦁  ByteLyon Mux (config)`)
	log.Log().Str("web.url", os.Getenv("WEB_URL")).Msg(`🦁 `)
	log.Log().Str("web.key", os.Getenv("WEB_KEY")).Msg(`🦁 `)
	log.Log().Str("s3.bucket", os.Getenv("S3_BUCKET")).Msg(`🦁 `)
	log.Log().Stringer("log.lvl", log.Logger.GetLevel()).Msg(`🦁 `)
	log.Log().Msgf("🦁\n")
}
