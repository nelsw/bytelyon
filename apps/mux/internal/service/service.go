package service

import (
	"fmt"

	"github.com/nelsw/bytelyon/apps/mux/internal/client"
	"github.com/nelsw/bytelyon/apps/mux/internal/config"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/rs/zerolog/log"
)

var broApiUrl, webApiUrl string

func init() {
	broApiUrl = fmt.Sprintf("%s/bots", config.Get("BRO_URL"))
	webApiUrl = fmt.Sprintf("%s/api/bots", config.Get("WEB_URL"))
}

func GetWebBots() model.Bots {
	bots := client.Get[model.Bots](webApiUrl)
	log.Info().
		Int("count", len(bots)).
		Msg("fetched bots from web api")
	return bots
}

func PostBroBot(b *model.Bot) {
	client.Post(broApiUrl, b)
	log.Info().
		EmbedObject(b).
		Msg("posted bot to bro api")
}
