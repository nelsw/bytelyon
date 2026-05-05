package shopify

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/rs/zerolog/log"
)

const shopifyAdmin = "https://msnbic-0w.myshopify.com/admin"
const shopifyAuth = shopifyAdmin + "/oauth/access_token"
const shopifyAPI = shopifyAdmin + "/api/2026-01/graphql.json"

type AccessTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type CreateArticleResponse struct {
	Data struct {
		ArticleCreate struct {
			Article    any `json:"article"`
			UserErrors []struct {
				Code    string   `json:"code"`
				Field   []string `json:"field"`
				Message string   `json:"message"`
			} `json:"userErrors"`
		} `json:"articleCreate"`
	} `json:"data"`
	Extensions struct {
		Cost struct {
			RequestedQueryCost int `json:"requestedQueryCost"`
			ActualQueryCost    int `json:"actualQueryCost"`
			ThrottleStatus     struct {
				MaximumAvailable   float64 `json:"maximumAvailable"`
				CurrentlyAvailable int     `json:"currentlyAvailable"`
				RestoreRate        float64 `json:"restoreRate"`
			} `json:"throttleStatus"`
		} `json:"cost"`
	} `json:"extensions"`
}

func (r *CreateArticleResponse) error() (err error) {
	for _, e := range r.Data.ArticleCreate.UserErrors {
		log.Warn().
			Str("code", e.Code).
			Strs("field", e.Field).
			Msg(e.Message)
		err = errors.Join(err, errors.New(e.Message))
	}
	return
}

func accessToken() (string, error) {

	out, err := https.PostForm(shopifyAuth, map[string][]string{
		"grant_type":    {"client_credentials"},
		"client_id":     {os.Getenv("SHOPIFY_CLIENT_ID")},
		"client_secret": {os.Getenv("SHOPIFY_CLIENT_SECRET")},
	})

	if err != nil {
		return "", err
	}

	var atr AccessTokenResponse
	if err = json.Unmarshal(out, &atr); err != nil {
		return "", err
	}

	return atr.AccessToken, nil
}

// CreateArticle creates a Shopify Article.
// https://shopify.dev/docs/api/admin-graphql/latest/mutations/articleCreate
func CreateArticle(in []byte) (err error) {

	var tkn string
	if tkn, err = accessToken(); err != nil {
		return
	}

	var out []byte
	out, err = https.PostJSON(shopifyAPI, in, map[string]string{
		"Content-Type":           "application/json",
		"X-Shopify-Access-Token": tkn,
	})

	if err != nil {
		return
	}

	var r CreateArticleResponse
	if err = json.Unmarshal(out, &r); err != nil {
		return
	} else if err = r.error(); err != nil {
		return
	}

	return nil
}
