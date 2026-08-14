package service

import (
	"github.com/nelsw/bytelyon/apps/mux/internal/client"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/rs/zerolog/log"
)

var BroApiUrl, WebApiUrl string

func GetWebBots() model.Bots {
	bots := client.Get[model.Bots](WebApiUrl)
	log.Info().
		Int("count", len(bots)).
		Msg("fetched bots from web api")
	return bots
}

func PostBroBot(b *model.Bot) {
	client.Post(BroApiUrl, b)
	log.Info().
		EmbedObject(b).
		Msg("posted bot to bro api")
}
