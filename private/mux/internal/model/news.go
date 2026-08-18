package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/apps/mux/internal/doc"
	"github.com/rs/zerolog/log"
)

type News struct {
	Body        string    `json:"body"`
	BotID       int       `json:"bot_id"`
	Description string    `json:"description"`
	ImgAlt      string    `json:"img_alt"`
	ImgUrl      string    `json:"img_url"`
	Keywords    []string  `json:"keywords"`
	PublishedAt time.Time `json:"published_at"`
	Publisher   string    `json:"publisher"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
}

func (n *News) Route() string {
	return fmt.Sprintf("/bots/%d/articles", n.BotID)
}

func (n *News) UnmarshalJSON(data []byte) (err error) {

	defer func() {
		if err != nil {
			log.Err(err).Bytes("data", data).Msg("News:UnmarshalJSON")
		}
	}()

	var v struct {
		BotID       int       `json:"bot_id"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		PublishedAt time.Time `json:"published_at"`
		Content     []byte    `json:"content"`
	}

	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}

	d := doc.New(v.Content, true)

	var publisher string
	if strings.HasPrefix(v.URL, "https://news.google") {
		publisher = "Google News"
	} else {
		publisher = "BingNews"
	}

	*n = News{
		Body:        d.Body(),
		BotID:       v.BotID,
		Description: d.Description(),
		ImgAlt:      d.ImgAlt(),
		ImgUrl:      d.ImgUrl(),
		Keywords:    d.Keywords(),
		PublishedAt: v.PublishedAt,
		Publisher:   publisher,
		Source:      d.Source(),
		Title:       v.Title,
		URL:         v.URL,
	}

	return
}
