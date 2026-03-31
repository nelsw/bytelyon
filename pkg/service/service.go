package service

import (
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/rs/zerolog/log"
)

func SpinArticle(a *model.Article) (link string, err error) {
	if a.Body, err = client.Prompt(a.Prompt, "Write a blog post summarizing this article: "+a.Body); err != nil {
		log.Warn().Err(err).Msg("failed to spin article html")
	} else if link, err = client.CreateArticle(a); err != nil {
		log.Error().Err(err).Msg("Failed to create article on Shopify after spinning it")
	} else {
		log.Info().Msg("published article")
	}
	return
}
