package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Fingerprint struct {
	// Cookies to set for context
	Cookies []playwright.Cookie `json:"cookies"`
	// Origins to set for context
	Origins []playwright.Origin `json:"origin"`
}

func NewFingerprint() *Fingerprint {
	return &Fingerprint{
		Cookies: []playwright.Cookie{},
		Origins: []playwright.Origin{},
	}
}

func (f *Fingerprint) String() string {
	b, _ := json.MarshalIndent(f, "", "\t")
	return string(b)
}

func (f *Fingerprint) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {

	var cookies []types.AttributeValue
	for _, c := range f.Cookies {
		var sameSite string
		if c.SameSite != nil {
			sameSite = string(*c.SameSite)
		}
		var partitionKey string
		if c.PartitionKey != nil {
			partitionKey = *c.PartitionKey
		}
		cookies = append(cookies, &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"name":         &types.AttributeValueMemberS{Value: c.Name},
				"value":        &types.AttributeValueMemberS{Value: c.Value},
				"domain":       &types.AttributeValueMemberS{Value: c.Domain},
				"path":         &types.AttributeValueMemberS{Value: c.Path},
				"expires":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", c.Expires)},
				"httpOnly":     &types.AttributeValueMemberBOOL{Value: c.HttpOnly},
				"secure":       &types.AttributeValueMemberBOOL{Value: c.Secure},
				"sameSite":     &types.AttributeValueMemberS{Value: sameSite},
				"partitionKey": &types.AttributeValueMemberS{Value: partitionKey},
			},
		})
	}

	var origins []types.AttributeValue
	for _, o := range f.Origins {
		var localStorage []types.AttributeValue
		for _, l := range o.LocalStorage {
			localStorage = append(localStorage, &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"name":  &types.AttributeValueMemberS{Value: l.Name},
					"value": &types.AttributeValueMemberS{Value: l.Value},
				},
			})
		}
		origins = append(origins, &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"origin":       &types.AttributeValueMemberS{Value: o.Origin},
				"localStorage": &types.AttributeValueMemberL{Value: localStorage},
			},
		})
	}

	return &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
		"cookies": &types.AttributeValueMemberL{Value: cookies},
		"origins": &types.AttributeValueMemberL{Value: origins},
	}}, nil
}

func (f *Fingerprint) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) error {
	var m map[string]types.AttributeValue
	if v.(*types.AttributeValueMemberM).Value == nil {
		return errors.New("bot unmarshal value was nil")
	}
	f.Cookies = []playwright.Cookie{}
	f.Origins = []playwright.Origin{}

	if val, ok := m["cookies"]; ok && val != nil {
		for _, cookieVal := range val.(*types.AttributeValueMemberL).Value {
			var c = playwright.Cookie{}
			cookieMap := cookieVal.(*types.AttributeValueMemberM).Value

			if cV, cK := cookieMap["name"]; cK {
				c.Name = cV.(*types.AttributeValueMemberS).Value
			}
			if cV, cK := cookieMap["value"]; cK {
				c.Value = cV.(*types.AttributeValueMemberS).Value
			}
			if cV, cK := cookieMap["domain"]; cK {
				c.Domain = cV.(*types.AttributeValueMemberS).Value
			}
			if cV, cK := cookieMap["path"]; cK {
				c.Path = cV.(*types.AttributeValueMemberS).Value
			}
			if cV, cK := cookieMap["expires"]; cK {
				exp, err := strconv.ParseFloat(cV.(*types.AttributeValueMemberN).Value, 64)
				if err != nil {
					log.Err(err).Msg("failed to parse cookie expiration")
				}
				c.Expires = exp
			}
			if cV, cK := cookieMap["httpOnly"]; cK {
				c.HttpOnly = cV.(*types.AttributeValueMemberBOOL).Value
			}
			if cV, cK := cookieMap["secure"]; cK {
				c.Secure = cV.(*types.AttributeValueMemberBOOL).Value
			}
			if cV, cK := cookieMap["sameSite"]; cK {
				c.SameSite = util.Ptr(playwright.SameSiteAttribute(cV.(*types.AttributeValueMemberS).Value))
			}
			if cV, cK := cookieMap["partitionKey"]; cK {
				c.PartitionKey = util.Ptr(cV.(*types.AttributeValueMemberS).Value)
			}

			f.Cookies = append(f.Cookies, c)
		}
	}

	if val, ok := m["origins"]; ok && val != nil {
		originMaps := val.(*types.AttributeValueMemberL).Value
		for _, originVal := range originMaps {
			var o = playwright.Origin{}
			originMap := originVal.(*types.AttributeValueMemberM).Value

			if oV, oK := originMap["origin"]; oK {
				o.Origin = oV.(*types.AttributeValueMemberS).Value
			}
			if oV, oK := originMap["localStorage"]; oK {
				for _, lsm := range oV.(*types.AttributeValueMemberL).Value {
					o.LocalStorage = append(o.LocalStorage, playwright.NameValue{
						Name:  lsm.(*types.AttributeValueMemberM).Value["name"].(*types.AttributeValueMemberS).Value,
						Value: lsm.(*types.AttributeValueMemberM).Value["value"].(*types.AttributeValueMemberS).Value,
					})
				}
			}
		}
	}

	//var cookies []types.AttributeValue
	//var origins []types.AttributeValue

	log.Debug().Msgf("UnmarshalDynamoDBAttributeValue: %+v", m)

	return nil
}

func (f *Fingerprint) SetState(s *playwright.StorageState) {
	f.Cookies = s.Cookies
	f.Origins = s.Origins
}

func (f *Fingerprint) GetState() *playwright.OptionalStorageState {
	var cookies []playwright.OptionalCookie
	for _, c := range f.Cookies {
		cookies = append(cookies, playwright.OptionalCookie{
			Name:         c.Name,
			Value:        c.Value,
			URL:          nil,
			Domain:       util.PtrOrNil(c.Domain),
			Path:         util.PtrOrNil(c.Path),
			Expires:      util.PtrOrNil(c.Expires),
			HttpOnly:     util.Ptr(c.HttpOnly),
			Secure:       util.Ptr(c.Secure),
			SameSite:     c.SameSite,
			PartitionKey: c.PartitionKey,
		})
	}
	return &playwright.OptionalStorageState{
		Origins: f.Origins,
		Cookies: cookies,
	}
}
