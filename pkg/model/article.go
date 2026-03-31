package model

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

const articleBlogID = "gid://shopify/Blog/82899828795"
const articleQuery = `mutation CreateArticle($article: ArticleCreateInput!) 
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
}`

type Article struct {
	ID      ulid.ULID
	Title   string
	Handle  string
	Body    string
	Summary string
	Tags    []string
	Image   map[string]string
	Prompt  string
}

func (a *Article) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if s, ok := m["title"]; !ok || s == "" {
		return errors.New("empty article title")
	} else {
		a.Title = s
	}
	if s, ok := m["body"]; !ok || s == "" {
		return errors.New("empty article body")
	} else {
		a.Body = s
	}
	if s, ok := m["prompt"]; !ok || s == "" {
		return errors.New("empty article prompt")
	} else {
		a.Prompt = s
	}

	if s, ok := m["id"]; ok {
		a.ID = ParseULID(s)
	} else {
		a.ID = NewULID()
	}

	a.Handle = strings.ToLower(strings.ReplaceAll(a.Title, " ", "-")) + "-" + a.ID.String()

	if s, ok := m["summary"]; ok {
		a.Summary = s
	}

	if s, ok := m["tags"]; ok {
		a.Tags = strings.Split(s, ",")
	}

	s, ok := m["image"]
	if !ok {
		return nil
	}
	if filepath.Ext(s) != ".png" {
		log.Warn().Str("image", s).Msg("image extension not supported")
		return nil
	}
	if strings.HasPrefix(s, "https://") {
		log.Warn().Str("image", s).Msg("url looks bad")
		return nil
	}

	a.Image = map[string]string{
		"altText": a.Title + " Image",
		"url":     s,
	}

	return nil
}

func (a *Article) ToShopifyPayload() []byte {

	b, _ := json.Marshal(map[string]any{
		"query": articleQuery,
		"variables": map[string]any{
			"article": map[string]any{
				"blogId":      articleBlogID,
				"title":       a.Title,
				"author":      map[string]any{"name": "Stu Andrew"},
				"handle":      strings.ToLower(strings.ReplaceAll(a.Title, " ", "-")) + "-" + a.ID.String(),
				"body":        a.Body,
				"summary":     a.Summary,
				"isPublished": true,
				"publishDate": a.ID.Timestamp(),
				"tags":        a.Tags,
				"image":       a.Image,
			},
		},
	})

	return b
}
