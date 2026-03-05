package model

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	// Either url or domain / path are required. Optional.
	URL *string `json:"url"`
	// For the cookie to apply to all subdomains as well, prefix domain with a dot, like this: ".example.com". Either url
	// or domain / path are required. Optional.
	Domain *string `json:"domain"`
	// Either url or domain / path are required Optional.
	Path *string `json:"path"`
	// Unix time in seconds. Optional.
	Expires *float64 `json:"expires"`
	// Optional.
	HttpOnly *bool `json:"httpOnly"`
	// Optional.
	Secure *bool `json:"secure"`
	// Optional.
	SameSite *string `json:"sameSite"`
}

type Storage struct {
	// Name of the header.
	Name string `json:"name"`
	// Value of the header.
	Value string `json:"value"`
}

type Origin struct {
	Origin string `json:"origin"`
	// LocalStorage to set for context
	LocalStorage []Storage `json:"localStorage"`
}

type BroCtxState struct {
	// Cookies to set for context
	Cookies []Cookie `json:"cookies"`
	Origins []Origin `json:"origins"`
}

func (bs BroCtxState) StorageState() *playwright.OptionalStorageState {
	b, _ := json.Marshal(&bs)
	var oss playwright.OptionalStorageState
	_ = json.Unmarshal(b, &oss)
	return &oss
}

func (bs BroCtxState) item() map[string]types.AttributeValue {
	i, err := attributevalue.MarshalMap(&bs)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal BroCtxState to DB item!")
		return map[string]types.AttributeValue{}
	}
	return i
}

func (bs *BroCtxState) unmarshal(i types.AttributeValue) error {
	return attributevalue.Unmarshal(i, bs)
}
