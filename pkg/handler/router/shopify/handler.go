package shopify

import (
	"encoding/json"
	"maps"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/service/images"
	"github.com/nelsw/bytelyon/pkg/shopify"
	"github.com/nelsw/bytelyon/pkg/store"
	"github.com/rs/zerolog/log"
)

var handleRegex = regexp.MustCompile("[^a-zA-Z0-9\\-]+")

func Handler(r Request) Response {

	r.Log()

	if r.IsGuest() {
		return r.NOPE()
	}

	switch r.Method() {
	case http.MethodPost:
		return handlePost(r)
	case http.MethodGet:
		return handleGet(r)
	}

	return r.NI()
}

func handlePost(r Request) Response {

	var p = new(model.Post)
	if err := json.Unmarshal([]byte(r.Body), p); err != nil {
		log.Err(err).Msg("failed to unmarshal post")
		return r.BAD(err)
	}

	// assign a new ID
	p.ID = model.NewULID()

	// define the handle
	p.Handle = strings.ToLower(strings.ReplaceAll(p.Title, " ", "-"))
	p.Handle = handleRegex.ReplaceAllString(p.Handle, "")
	for strings.Contains(p.Handle, "--") {
		p.Handle = strings.ReplaceAll(p.Handle, "--", "-")
	}
	p.Handle += "-" + p.ID.String()

	// convert the image url to a public url of type .png
	if p.ImgSrc != "" {
		if url, err := images.ToPublicURL(p.ImgSrc); err != nil {
			log.Warn().Err(err).Msgf("Failed to convert url to public url")
			p.ImgSrc = ""
			p.ImgAlt = ""
		} else {
			p.ImgSrc = url
		}
		if p.ImgAlt == "" {
			p.ImgAlt = p.Title + " image"
		}
	}

	tkn, err := shopify.AccessToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Shopify access token")
		return r.BAD(err)
	}

	_, err = shopify.PostArticle(
		tkn,
		os.Getenv("SHOPIFY_STORE"),
		"gid://shopify/Blog/82899828795",
		p.Title,
		"Stu Andrew",
		p.Handle,
		p.Body,
		p.Summary,
		p.ImgSrc,
		p.ImgAlt,
		p.PublishedAt,
		p.Tags,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create article on Shopify after spinning it")
		return r.BAD(err)
	}

	return r.OK(map[string]any{"link": "https://firefibers.com/blogs/news/" + p.Handle})
}

func handleGet(r Request) Response {

	orderDB, err := store.New[string, shopify.Order]("orders.json")
	if err != nil {
		return r.BAD(err)
	}
	orderDB.Close()

	var orders []any
	customers := map[string]any{}
	for _, order := range orderDB.All() {
		orders = append(orders, order.Row())
		c := order.Customer
		c.Ordered = order.CreatedAt
		if _, ok := customers[order.Customer.ID]; !ok {
			customers[c.ID] = c.Row()
			continue
		}
		a, _ := time.Parse(time.RFC3339, c.Ordered)
		b, _ := time.Parse(time.RFC3339, customers[c.ID].(map[string]any)["ordered"].(string))
		if a.After(b) {
			customers[c.ID] = c.Row()
		}
	}

	return r.OK(map[string]any{
		"customers": slices.Collect(maps.Values(customers)),
		"orders":    orders,
	})
}
