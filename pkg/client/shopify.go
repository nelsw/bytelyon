package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
)

type AccessTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func AccessToken() (string, error) {
	res, err := http.PostForm("https://msnbic-0w.myshopify.com/admin/oauth/access_token", url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{os.Getenv("SHOPIFY_CLIENT_ID")},
		"client_secret": []string{os.Getenv("SHOPIFY_CLIENT_SECRET")},
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

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

// https://shopify.dev/docs/api/admin-graphql/latest/mutations/articleCreate
func CreateArticle() {

}
