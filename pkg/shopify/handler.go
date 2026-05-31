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

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/image"
	"github.com/nelsw/bytelyon/pkg/store"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var handleRegex = regexp.MustCompile(`[^a-zA-Z0-9\\-]+`)

type Post struct {
	Body        string       `json:"body"`
	Handle      string       `json:"handle"`
	ID          ulid.ULID    `json:"id"`
	Image       *image.Model `json:"image"`
	PublishedAt time.Time    `json:"publishedAt"`
	Summary     string       `json:"summary"`
	Tags        []string     `json:"tags"`
	Title       string       `json:"title"`
}

func Handler(r api.HTTPRequest) api.HTTPResponse {

	if r.IsGuest() {
		return api.Forbidden()
	}

	switch r.RequestContext.HTTP.Method {
	case http.MethodPost:
		return handlePost(r)
	case http.MethodGet:
		return handleGet()
	}

	return api.NotImplemented()
}

func handlePost(r api.HTTPRequest) api.HTTPResponse {

	var p = new(Post)
	if err := json.Unmarshal([]byte(r.Body), p); err != nil {
		log.Err(err).Msg("failed to unmarshal post")
		return api.BadRequest(err)
	}

	// assign a new ID
	p.ID = id.NewULID()

	// define the handle
	p.Handle = strings.ToLower(strings.ReplaceAll(p.Title, " ", "-"))
	p.Handle = handleRegex.ReplaceAllString(p.Handle, "")
	for strings.Contains(p.Handle, "--") {
		p.Handle = strings.ReplaceAll(p.Handle, "--", "-")
	}
	p.Handle += "-" + p.ID.String()

	// convert the image url to a public url of type .png
	if p.Image == nil || !p.Image.ConvertToPNG() {
		p.Image = new(image.Model)
	} else if p.Image.URL != "" && p.Image.ALT == "" {
		p.Image.ALT = p.Title + " image"
	}

	tkn, err := AccessToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Shopify access token")
		return api.BadRequest(err)
	}

	_, err = PostArticle(
		tkn,
		os.Getenv("SHOPIFY_STORE"),
		"gid://shopify/Blog/82899828795",
		p.Title,
		"Stu Andrew",
		p.Handle,
		p.Body,
		p.Summary,
		p.Image,
		p.PublishedAt,
		p.Tags,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create article on Shopify after spinning it")
		return api.BadRequest(err)
	}

	return api.OK(map[string]any{"link": "https://firefibers.com/blogs/news/" + p.Handle})
}

func handleGet() api.HTTPResponse {

	orderDB, err := store.New[string, Order]("orders.json")
	if err != nil {
		return api.BadRequest(err)
	}
	defer func() {
		if closeErr := orderDB.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("failed to close orderDB")
		}
	}()

	var orders []any
	customers := map[string]any{}
	for _, order := range orderDB.Values() {
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
	v := make([]any, 0, len(customers))
	for _, c := range slices.Sorted(maps.Keys(customers)) {
		v = append(v, customers[c])
	}
	return api.OK(map[string]any{
		"customers": v,
		"orders":    orders,
	})
}
