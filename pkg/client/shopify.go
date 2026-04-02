package client

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/nelsw/bytelyon/pkg/model"
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
	res, err := http.PostForm(shopifyAuth, url.Values{
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

// CreateArticle creates a Shopify Article.
// https://shopify.dev/docs/api/admin-graphql/latest/mutations/articleCreate
func CreateArticle(a *model.Article) (s string, err error) {

	var tkn string
	if tkn, err = accessToken(); err != nil {
		return
	}

	var b []byte
	b, err = PostJSON(shopifyAPI, a.ToShopifyPayload(), map[string]string{
		"Content-Type":           "application/json",
		"X-Shopify-Access-Token": tkn,
	})

	if err != nil {
		return
	}

	log.Debug().Str("body", string(b)).Msg("shopify response")

	var r CreateArticleResponse
	if err = json.Unmarshal(b, &r); err != nil {
		return
	} else if err = r.error(); err != nil {
		return
	}

	return a.GetLink(), nil
}
