package articles

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nelsw/bytelyon/pkg/ai"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/service/images"
	"github.com/nelsw/bytelyon/pkg/shopify"
	"github.com/rs/zerolog/log"
)

func Handler(r api.Request) api.Response {
	if r.IsGuest() {
		return r.NOPE()
	}
	switch r.Method() {
	case http.MethodPost:
		return handlePost(r)
	case http.MethodPut:
		return handlePut(r)
	}
	return r.NI()
}

func handlePost(r api.Request) api.Response {
	var a = new(model.Article)
	if err := json.Unmarshal([]byte(r.Body), a); err != nil {
		log.Err(err).Msg("failed to unmarshal article")
		return r.BAD(err)
	}

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

	if txt, err := ai.Prompt("you", message, true); err != nil {
		log.Err(err).Msg("failed to spin article body")
	} else {
		a.Body = txt
	}

	err := shopify.CreateArticle(a.ToShopifyPayload())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create article on Shopify after spinning it")
		return r.BAD(err)
	}

	return r.OK(map[string]any{"link": "https://firefibers.com/blogs/news/" + a.Handle})
}

func handlePut(r api.Request) api.Response {
	var a = new(model.Article)
	if err := json.Unmarshal([]byte(r.Body), a); err != nil {
		log.Err(err).Msg("failed to unmarshal article")
		return r.BAD(err)
	}

	err := shopify.CreateArticle(a.ToShopifyPayload())
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(map[string]any{"link": "https://firefibers.com/blogs/news/" + a.Handle})
}
