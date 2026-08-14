package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var vars map[string]string

func init() {

	var err error
	if vars, err = godotenv.Read(); err != nil {
		panic(err)
	}

	var lvl zerolog.Level
	if lvl, err = zerolog.ParseLevel(os.Getenv("LOG_LEVEL")); err != nil {
		panic(err)
	}

	log.Logger = MakeLogger(lvl, nil)
}

func Get(key string) string { return vars[key] }

func Print() {
	log.Log().Msgf("\n🦁")
	log.Log().Msg(`🦁  ByteLyon Mux (config)`)
	for k, v := range vars {
		log.Log().Any(k, v).Msg(`🦁 `)
	}
	log.Log().Msgf("🦁\n")
}
