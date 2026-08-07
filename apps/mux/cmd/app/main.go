package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nelsw/bytelyon/apps/mux/pkg/logger"
	"github.com/nelsw/bytelyon/apps/mux/pkg/manager"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var host, key, lvl string

	flag.StringVar(&host, "web", "http://localhost", "web app (web) host name")
	flag.StringVar(&key, "key", "my-random-32-character-x-web-key", "web app (web) auth key")
	flag.StringVar(&lvl, "log", zerolog.TraceLevel.String(), "log level trace->disabled")
	flag.Parse()

	log.Logger = *logger.New(lvl)

	log.Log().Msgf("\n🦁")
	log.Log().Msg(`🦁  ByteLyon Muxer`)
	log.Log().Str("web", host).Msg(`🦁 `)
	log.Log().Str("key", "..."+key[26:]).Msg(`🦁 `)
	log.Log().Str("log", lvl).Msg(`🦁 `)
	log.Log().Msgf("🦁\n")

	mgr := manager.New(8085, host, key)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /bots", mgr.WebHandler())
	mux.HandleFunc("PUT /bots", mgr.BroHandler())

	fmt.Println("Server starting on :3000...")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		fmt.Printf("Server failed: %s\n", err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	mgr.Quit()
	fmt.Printf("\n👋\n")
}
