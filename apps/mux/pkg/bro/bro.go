package bro

import (
	"errors"
	"fmt"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/pkg/http"
	"github.com/nelsw/bytelyon/apps/mux/pkg/model"
)

type Client struct {
	port int
}

func New(port int) *Client {
	return &Client{port}
}

func (c *Client) Put(b *model.Bot) (err error) {
	if b == nil {
		return errors.New("bot is nil")
	}

	url := fmt.Sprintf("http://localhost:%d/%s/%d/%s", c.port, b.Type, b.ID, b.Query)

	switch b.Type {
	case model.NewsBot:
		if !b.Since.IsZero() {
			url += "/" + b.Since.Format(time.RFC3339)
		}
		_, err = http.Put(url, nil, nil)
	case model.SearchBot:
		_, err = http.Put(url, nil, nil)
	case model.SitemapBot:
		_, err = http.Put(url, nil, nil)
	default:
		err = errors.New("unknown bot type")
	}
	return
}
