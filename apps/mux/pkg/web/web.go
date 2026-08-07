package web

import (
	"fmt"

	"github.com/nelsw/bytelyon/apps/mux/pkg/http"
	"github.com/nelsw/bytelyon/apps/mux/pkg/model"
	"github.com/rs/zerolog/log"
)

type Client struct {
	host, key string
}

func New(host, key string) *Client {
	return &Client{host, key}
}

func (c *Client) url(route string) string {
	return c.host + "/api/" + route
}

func (c *Client) header() http.Header {
	return http.Header{"x-api-key": []string{c.key}}
}

func (c *Client) GetBots() (bots model.Bots, err error) {
	if bots, err = http.Get[model.Bots](c.url("bots"), c.header()); err != nil {
		log.Err(err).Msg("failed to get bots")
	} else {
		log.Info().Msgf("[%d] bots ready", len(bots))
	}
	return
}

func (c *Client) PutBot(bot *model.Bot, err error) {
	route := fmt.Sprintf("bots/%d", bot.ID)
	data := map[string]any{"result": "ok"}
	if err != nil {
		data["result"] = err.Error()
	}
	if _, err = http.Put(c.url(route), data, c.header()); err != nil {
		log.Err(err).EmbedObject(bot).Msg("failed to put bot")
	}
}

func (c *Client) PutSitemap(bot *model.Bot, sitemap model.Sitemap) {
	if _, err := http.Put(fmt.Sprintf("bots/%d/sitemaps", bot.ID), sitemap, c.header()); err != nil {
		log.Err(err).EmbedObject(bot).Any("sitemap", sitemap).Msg("failed to put sitemap")
	}
}

func (c *Client) PutSerp(bot *model.Bot, serp model.Serp) {
	if _, err := http.Put(fmt.Sprintf("bots/%d/searches", bot.ID), serp, c.header()); err != nil {
		log.Err(err).EmbedObject(bot).Any("serp", serp).Msg("failed to put serp")
	}
}

func (c *Client) PutPage(bot *model.Bot, page model.Page) {
	var err error
	if bot.Type == model.SearchBot {
		_, err = http.Put(fmt.Sprintf("searches/%d/page", bot.SerpID), page, c.header())
	} else {
		_, err = http.Put(fmt.Sprintf("sitemaps/%d/page", bot.SitemapID), page, c.header())
	}
	if err != nil {
		log.Err(err).EmbedObject(bot).Msgf("failed to create %s page %v", bot.Type, page)
	}
}

func (c *Client) PutArticles(bot *model.Bot, arr []any) {
	for _, v := range arr {
		if _, err := http.Put(fmt.Sprintf("bots/%d/articles", bot.ID), v, c.header()); err != nil {
			log.Err(err).EmbedObject(bot).Msgf("failed to create article %v", v)
		}
	}
}
