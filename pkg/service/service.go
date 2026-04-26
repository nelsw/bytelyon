package service

import (
	"strings"

	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/service/images"
	"github.com/rs/zerolog/log"
)

func SpinArticle(a *model.Article) (string, error) {

	if img, ok := a.Image["url"]; ok {
		if url, err := images.ToPublicURL(img); err != nil {
			log.Warn().Err(err).Msgf("Failed to convert url to public url")
			a.Image = nil
		} else {
			a.Image["url"] = url
		}
	}

	message := "Write a blog post summarizing this article: " + a.Body
	if len(a.Keywords) > 0 {
		message += `\n\ninclude the following keywords: ` + strings.Join(a.Keywords, ",") + `.`
	}
	if a.URL != "" {
		message += `\n\ninclude a link to the original article: ` + a.URL
	}

	if txt, err := client.Prompt(a.Prompt, message, true); err != nil {
		log.Err(err).Msg("failed to spin article body")
	} else {
		a.Body = txt
	}

	link, err := client.CreateArticle(a)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create article on Shopify after spinning it")
		return "", err
	}

	log.Info().Msg("published article")
	return link, nil
}
