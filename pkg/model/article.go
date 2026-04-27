package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var handleRegex = regexp.MustCompile("[^a-zA-Z0-9\\-]+")

type Article struct {
	ID          ulid.ULID
	Title       string
	Handle      string
	Body        string
	Summary     string
	Tags        []any
	Image       map[string]string
	ImageAlt    string
	Prompt      string
	PublishedAt time.Time
	URL         string
	Keywords    []string
}

func (a *Article) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if s, ok := m["title"]; !ok || s == "" {
		return errors.New("empty article title")
	} else {
		a.Title = s.(string)
	}
	if s, ok := m["body"]; !ok || s == "" {
		return errors.New("empty article body")
	} else {
		a.Body = s.(string)
	}
	if s, ok := m["prompt"]; !ok || s == "" {
		return errors.New("empty article prompt")
	} else {
		a.Prompt = s.(string)
	}

	a.PublishedAt = time.Now()
	if s, ok := m["publishedAt"]; ok {
		if t, err := time.Parse("2006-01-02 15:04", s.(string)); err != nil {
			log.Warn().Err(err).Msg("failed to parse published at")
		} else {
			a.PublishedAt = t
		}
	}
	a.ID = NewULID(a.PublishedAt)

	a.setHandle()

	if s, ok := m["url"]; ok {
		a.URL = s.(string)
	}

	if s, ok := m["keywords"]; ok {
		for _, v := range s.([]any) {
			a.Keywords = append(a.Keywords, fmt.Sprintf("%v", v))
		}
	}

	if s, ok := m["summary"]; ok {
		a.Summary = s.(string)
	}

	if v, ok := m["tags"]; ok {
		if _, ok = v.([]any); ok {
			a.Tags = v.([]any)
		} else if _, ok = v.(string); ok {
			a.Tags = []any{v.(string)}
		}
	}

	s, ok := m["image"]
	if !ok {
		log.Warn().Msg("article has no image")
		return nil
	}
	if !strings.HasPrefix(s.(string), "https://") {
		log.Warn().Str("image", s.(string)).Msg("url looks bad")
		return nil
	}

	if a.ImageAlt == "" {
		a.ImageAlt = a.Title + " Image"
	}

	a.Image = map[string]string{
		"altText": a.ImageAlt,
		"url":     s.(string),
	}

	return nil
}

func (a *Article) ToShopifyPayload() []byte {

	a.setHandle()

	b, _ := json.Marshal(map[string]any{
		"query": `mutation CreateArticle($article: ArticleCreateInput!) 
{ 
	articleCreate(article: $article) { 
		article { 
			id 
			title 
			author { name } 
			handle 
			body 
			summary 
			tags 
			image { altText originalSrc } 
		} 
		userErrors { code field message } 
	} 
}`,
		"variables": map[string]any{
			"article": map[string]any{
				"blogId":      "gid://shopify/Blog/82899828795",
				"title":       a.Title,
				"author":      map[string]any{"name": "Stu Andrew"},
				"handle":      a.Handle,
				"body":        a.Body,
				"summary":     a.Summary,
				"isPublished": true,
				"publishDate": a.PublishedAt,
				"tags":        a.Tags,
				"image":       a.Image,
			},
		},
	})

	return b
}

func (a *Article) GetLink() string {
	return "https://firefibers.com/blogs/news/" + a.Handle
}

func (a *Article) setHandle() {

	// check if the article title is in the body
	if z := strings.Index(a.Body, `</h1>`); strings.Index(a.Body, `<h1>`) == 0 && z != -1 {
		a.Title = a.Body[4:z]
	}

	// replace all spaces with dashes
	a.Handle = strings.ToLower(strings.ReplaceAll(a.Title, " ", "-"))

	// remove all non-alphanumeric and non-dash characters
	a.Handle = handleRegex.ReplaceAllString(a.Handle, "")

	// remove duplicate dashes
	for strings.Contains(a.Handle, "--") {
		a.Handle = strings.ReplaceAll(a.Handle, "--", "-")
	}

	// append with the article id
	a.Handle += "-" + a.ID.String()
}
