package model

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

type Meta struct {
	Abstract               string `meta:"abstract" json:"abstract,omitempty"`
	Description            string `meta:"description" json:"description,omitempty"`
	Google                 string `meta:"google" json:"google,omitempty"`
	GoogleSiteVerification string `meta:"google-site-verification" json:"googleSiteVerification,omitempty"`
	Headline               string `meta:"headline" json:"headline,omitempty"`
	Image                  string `meta:"image" json:"image,omitempty"`
	Keywords               string `meta:"keywords" json:"keywords,omitempty"`
	MsApplicationConfig    string `meta:"msapplication-config" json:"msapplicationConfig,omitempty"`
	NewsKeywords           string `meta:"news_keywords" json:"newsKeywords,omitempty"`
	Referrer               string `meta:"referrer" json:"referrer,omitempty"`
	Robots                 string `meta:"robots" json:"robots,omitempty"`
	Title                  string `meta:"title" json:"title,omitempty"`
	ThemeColor             string `meta:"theme-color" json:"themeColor,omitempty"`
	Thumbnail              string `meta:"thumbnail" json:"thumbnail,omitempty"`
	Viewport               string `meta:"viewport" json:"viewport,omitempty"`

	ArticleSection string `meta:"article:section" json:"articleAection,omitempty"`

	OgArticleAuthor         string `meta:"og:article:author" json:"ogArticleAuthor,omitempty"`
	OgArticleExpirationTime string `meta:"og:article:expiration_time" json:"ogArticleExpirationTime,omitempty"`
	OgArticleModifiedTime   string `meta:"og:article:modified_time" json:"ogArticleModifiedTime,omitempty"`
	OgArticlePublishedTime  string `meta:"og:article:published_time" json:"ogArticlePublishedTime,omitempty"`
	OgArticleSection        string `meta:"og:article:section" json:"ogArticleSection,omitempty"`
	OgArticleTag            string `meta:"og:article:tag" json:"ogArticleTag,omitempty"`
	OgDescription           string `meta:"og:description" json:"ogDescription,omitempty"`
	OgImage                 string `meta:"og:image" json:"ogImage,omitempty"`
	OgImageAlt              string `meta:"og:image:alt" json:"ogImageAlt,omitempty"`
	OgImageHeight           string `meta:"og:image:height" json:"ogImageHeight,omitempty"`
	OgImageSecureUrl        string `meta:"og:image:secure_url" json:"ogImageSecureUrl,omitempty"`
	OgImageUrl              string `meta:"og:image:url" json:"ogImageUrl,omitempty"`
	OgImageWidth            string `meta:"og:image:width" json:"ogImageWidth,omitempty"`
	OgLocale                string `meta:"og:locale" json:"ogLocale,omitempty"`
	OgPriceAmount           string `meta:"og:price:amount" json:"ogPriceAmount,omitempty"`
	OgPriceCurrency         string `meta:"og:price:currency" json:"ogPriceCurrency,omitempty"`
	OgSite                  string `meta:"og:site" json:"ogSite,omitempty"`
	OgSiteName              string `meta:"og:site_name" json:"ogSiteName,omitempty"`
	OgTitle                 string `meta:"og:title" json:"ogTitle,omitempty"`
	OgType                  string `meta:"og:type" json:"ogType,omitempty"`
	OgUrl                   string `meta:"og:url" json:"ogUrl,omitempty"`

	ProductAvailability  string `meta:"product:availability" json:"productAvailability,omitempty"`
	ProductPriceAmount   string `meta:"product:price:amount" json:"productPriceAmount,omitempty"`
	ProductPriceCurrency string `meta:"product:price:currency" json:"productPriceCurrency,omitempty"`

	ShopifyCheckoutApiToken string `meta:"shopify-checkout-api-token" json:"shopifyCheckoutApiToken,omitempty"`
	ShopifyDigitalWallet    string `meta:"shopify-digital-wallet" json:"shopifyDigitalWallet,omitempty"`

	TwitterCard        string `meta:"twitter:card" json:"twitterCard,omitempty"`
	TwitterDescription string `meta:"twitter:description" json:"twitterDescription,omitempty"`
	TwitterImage       string `meta:"twitter:image" json:"twitterImage,omitempty"`
	TwitterImageAlt    string `meta:"twitter:image:alt" json:"twitterImageAlt,omitempty"`
	TwitterImageSrc    string `meta:"twitter:image:src" json:"twitterImageSrc,omitempty"`
	TwitterSite        string `meta:"twitter:site" json:"twitterSite,omitempty"`
	TwitterTitle       string `meta:"twitter:title" json:"twitterTitle,omitempty"`
}

func ParseMeta(content string) Meta {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		log.Err(err).Msg("failed to parse meta")
		return Meta{}
	}
	m := map[string]string{}

	var key, val string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		key, val = "", ""
		if key = s.AttrOr("name", ""); key == "" {
			if key = s.AttrOr("property", ""); key == "" {
				return
			}
		}
		if val = s.AttrOr("content", ""); val == "" {
			return
		}
		m[key] = val
	})
	return MakeMeta(m)
}

func MakeMeta(m map[string]string) Meta {

	var out = new(Meta)

	in := m
	t := reflect.TypeOf(out).Elem()
	v := reflect.ValueOf(out).Elem()

	for i := 0; i < t.NumField(); i++ {
		k := t.Field(i).Tag.Get("meta")
		if val, ok := in[k]; ok {
			v.Field(i).SetString(val)
			delete(in, k)
		}
	}

	if len(in) > 0 {
		log.Trace().Msgf("unknown meta tags: %v", string(util.JSON(in)))
	}

	return *out
}

func (m Meta) String() string { return string(util.JSON(m)) }

func (m Meta) Img() Image {
	return MakeImage(m.ImageSrc(), m.ImageAlt())
}

func (m Meta) ImageSrc() string {
	return util.OrFunc(
		func(or string) bool { return strings.HasPrefix(or, "https") },
		m.Image,
		m.OgImage,
		m.OgImageUrl,
		m.OgImageSecureUrl,
		m.TwitterImage,
		m.TwitterImageSrc,
	)
}

func (m Meta) ImageAlt() string {
	return util.Or(
		m.OgImageAlt,
		m.TwitterImageAlt,
	)
}

func (m Meta) Source() string {
	return util.Or(
		m.OgSite,
		m.OgSiteName,
		m.TwitterSite,
	)
}

func (m Meta) Desc() string {
	return util.Or(
		m.Abstract,
		m.Description,
		m.OgDescription,
		m.TwitterDescription,
	)
}

func (m Meta) Keywerds() []string {

	opts := []string{
		m.Keywords,
		m.NewsKeywords,
	}

	kw := make(map[string]bool)
	for _, opt := range opts {
		if opt == "" {
			continue
		}
		kws := strings.Split(opt, ",")
		for _, w := range kws {
			kw[strings.TrimSpace(w)] = true
		}
	}

	if len(kw) == 0 {
		return []string{}
	}

	return slices.Sorted(maps.Keys(kw))
}

func (m Meta) PublishedAt() time.Time {
	var t time.Time
	t, _ = time.Parse(time.RFC3339, m.OgArticlePublishedTime)
	return t
}
