package shopify

import (
	"errors"
	"os"
)

var ByteLyon = &Credentials{
	Store: os.Getenv("SHOPIFY_CREDENTIALS"),
	Client: struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}{os.Getenv("SHOPIFY_CLIENT_ID"), os.Getenv("SHOPIFY_SECRET")},
}

type Credentials struct {
	Store  string `json:"store"`
	Client struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	} `json:"Client"`
}

//func (c *Credentials) adminApi() string {
//	return fmt.Sprintf("https://%s.myshopify.com/admin/api/2026-01/graphql.json", c.Store)
//}
//
//func (c *Credentials) accessToken() (tkn string, err error) {
//
//	if err = c.Validate(); err != nil {
//		return
//	}
//
//	var out []byte
//	if out, err = https.PostForm(
//		fmt.Sprintf("https://%s.myshopify.com/admin/oauth/access_token", c.Store),
//		map[string][]string{
//			"grant_type":    {"client_credentials"},
//			"client_id":     {c.Client.ID},
//			"client_secret": {c.Client.Secret},
//		}); err != nil {
//		return
//	}
//
//	var atr struct {
//		AccessToken string `json:"access_token"`
//	}
//
//	if err = json.Unmarshal(out, &atr); err == nil {
//		return
//	}
//
//	return atr.AccessToken, nil
//}

func (c *Credentials) Blogs() (map[string]string, error) {
	// todo - client ƒ
	// todo - handler f
	return nil, nil
}

func (c *Credentials) Validate() (err error) {
	if c.Store == "" {
		err = errors.Join(err, errors.New("store is required"))
	}
	if c.Client.ID == "" {
		err = errors.Join(err, errors.New("client id is required"))
	}
	if c.Client.Secret == "" {
		err = errors.Join(err, errors.New("client secret is required"))
	}
	return
}
