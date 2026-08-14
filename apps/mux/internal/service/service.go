package service

import (
	"fmt"
	"net/http"

	"github.com/nelsw/bytelyon/apps/mux/internal/client"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	WebHdr http.Header
	WebUrl string
)

func GetBots() model.Bots {
	bots := client.Get[model.Bots](WebUrl, WebHdr)
	log.Info().
		Int("count", len(bots)).
		Msg("fetched bots from web api")
	return bots
}

func PutBot(b *model.Bot, a any) {
	url := fmt.Sprintf("%s/bots/%d", WebUrl, b.ID)
	client.Put(url, WebHdr, a)
}

func PostJob(b *model.Bot) {
	client.Post("http://localhost:8085/bots", b)
	log.Info().
		EmbedObject(b).
		Msg("posted bot to bro api")
}
