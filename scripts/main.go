package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

var req *http.Request

func main() {

    var key string
    var poll, full, less int
	flag.StringVar(&key, "key", "my-random-32-character-x-api-key", "API Auth Key")
	flag.IntVar(&poll, "poll", 5, "Polling Interval in seconds")
	flag.IntVar(&full, "full", 1, "Number of full workers")
	flag.IntVar(&less, "less", 3, "Number of less workers")
	flag.Parse()
	
	if len(key) == 36 {
	    req, _ = http.NewRequest("GET", "https://bytelyon.com/api/bots", nil)
	} else {
	    req, _ = http.NewRequest("GET", "http://localhost:80/api/bots", nil)
	}
	req.Header.Set("x-api-key", key)

	log.Logger = makeLogger()

	log.Log().Msg(`🦁 `)
	log.Log().Msg(`🦁 ByteLyon Bot Runner`)
	log.Log().Str("api", req.RequestURI).Msg(`🦁 `)
	log.Log().Msg(`🦁 `)
    
    q := Start(less, full, poll)
    
    quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	q.Quit()

	log.Info().Msgf("\n👋") // newline for ^C buffer entry
}

