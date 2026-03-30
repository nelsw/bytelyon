package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

const shopifyAdmin = "https://msnbic-0w.myshopify.com/admin"
const shopifyAdminOAuth = shopifyAdmin + "/oauth/access_token"
const shopifyAdminGraphQL = shopifyAdmin + "/api/2026-01/graphql.json"
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

type AccessTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func accessToken() (string, error) {
	res, err := http.PostForm(shopifyAdminOAuth, url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{os.Getenv("SHOPIFY_CLIENT_ID")},
		"client_secret": []string{os.Getenv("SHOPIFY_CLIENT_SECRET")},
	})
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var b []byte
	if b, err = io.ReadAll(res.Body); err != nil {
		return "", err
	}

	var atr AccessTokenResponse
	if err = json.Unmarshal(b, &atr); err != nil {
		return "", err
	}

	return atr.AccessToken, nil
}

// CreateArticle creates a Shopify Article ... dingus.
// https://shopify.dev/docs/api/admin-graphql/latest/mutations/articleCreate
func CreateArticle(id ulid.ULID, title, body, ts, img string) (s string, err error) {

	var tkn string
	if tkn, err = accessToken(); err != nil {
		return
	}

	handle := strings.ToLower(strings.ReplaceAll(title, " ", "-")) + "-" + id.String()

	b, _ := json.Marshal(map[string]any{
		"query": articleQuery,
		"variables": map[string]any{
			"article": map[string]any{
				"blogId":      "gid://shopify/Blog/82899828795",
				"title":       title,
				"author":      map[string]any{"name": "Stu Andrew"},
				"handle":      handle,
				"body":        body,
				"summary":     "",
				"isPublished": true,
				"publishDate": ts,
				"tags":        []string{},
				"image": map[string]any{
					"altText": title + " Image",
					"url":     img,
				},
			},
		},
	})

	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, shopifyAdminGraphQL, bytes.NewBuffer(b)); err != nil {
		log.Err(err).Msg("failed to create article request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", tkn)

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		log.Err(err).Msg("Error creating Shopify article")
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var out []byte
	if out, err = io.ReadAll(res.Body); err != nil {
		return
	}

	log.Info().Bytes("body", out).Msg("create article response")

	if res.StatusCode != http.StatusOK {
		err = errors.New(string(out))
		log.Err(err).Msg("Error creating article")
		return
	}

	log.Info().
		Int("status", res.StatusCode).
		Msg("Shopify article created")

	return "https://firefibers.com/blogs/news/" + handle, nil
}
