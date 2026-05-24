package model

import (
	"fmt"
	"regexp"
	"testing"
	"unicode"

	"github.com/nelsw/bytelyon/pkg/util"
)

func TestMeta(t *testing.T) {
	m := map[string]string{
		"description":               "Modern lithium fire protection for businesses, fire departments, and distributors. Certified ISO, EN, and NFPA solutions for safer, compliant operations.",
		"og:description":            "Modern lithium fire protection for businesses, fire departments, and distributors. Certified ISO, EN, and NFPA solutions for safer, compliant operations.",
		"og:image":                  "http://li-fire.com/cdn/shop/files/logo_85395cb8-f7f3-42c9-ad5c-0ac898f862a2.png?v=1738541386",
		"og:image:height":           "680",
		"og:article:published_time": "2024-07-22T15:00:00-04:00",
		"google-site-verification":  "a;dlfka;dlfkjasd;fjadsa;sldkjfads;ljf",
		"og:image:secure:urls":      "https://li-fire.com/cdn/shop/files/logo_85395cb8-f7f3-42c9-ad5c-0ac898f862a2.png?v=1738541386",
		"shopify:digital:wallet":    "/69840568537/digital_wallets/dialog",
		"twitter:card":              "summary_large_image",
	}

	notAlphaNum := regexp.MustCompile(`[^a-zA-Z0-9]`)

	meta := make(map[string]string)

	for k, v := range m {

		for si := notAlphaNum.FindStringIndex(k); len(si) == 2; si = notAlphaNum.FindStringIndex(k) {
			l := k[:si[0]]
			c := string(unicode.ToUpper(rune(k[si[1]])))
			r := k[si[1]+1:]
			k = l + c + r
		}
		meta[k] = v
	}
	fmt.Println(string(util.JSON(meta)))
}
