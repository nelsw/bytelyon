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

func AccessToken() (string, error) {
	return accessToken()
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
