package client

import (
	"bytes"
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

// CreateArticle creates a Shopify Article ... dingus.
// https://shopify.dev/docs/api/admin-graphql/latest/mutations/articleCreate
func CreateArticle(a *model.Article) (s string, err error) {

	var tkn string
	if tkn, err = accessToken(); err != nil {
		return
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, shopifyAPI, bytes.NewBuffer(a.ToShopifyPayload())); err != nil {
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

	return "https://firefibers.com/blogs/news/" + a.Handle, nil
}
