package profile

import (
	"github.com/nelsw/bytelyon/pkg/anthropic"
	"github.com/nelsw/bytelyon/pkg/shopify"
)

type Model struct {

	// Anthropic API credentials
	Anthropic anthropic.Credentials

	// Img is a URL for a user avatar
	Img string `json:"img"`

	// Shopify API credentials
	Shopify shopify.Credentials `json:"shopify"`

	// Verified flag for email confirmation
	Verified bool `json:"verified"`
}
