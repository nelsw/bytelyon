package shopify

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/pkg/https"
)

type ErrorResponse struct {
	Errors []struct {
		Message   string `json:"message"`
		Locations []struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		}
		Path       []string `json:"path"`
		Extensions struct {
			Code         string `json:"code"`
			VariableName string `json:"variableName"`
		} `json:"extensions"`
	} `json:"errors"`
}

type Graph[T any] struct {
	Edges []struct {
		Node T `json:"node"`
	} `json:"edges"`
}

func (g Graph[T]) Slice() []T {
	var arr []T
	for _, e := range g.Edges {
		arr = append(arr, e.Node)
	}
	return arr
}

type Error struct {
	Errors []struct {
		Message   string `json:"message"`
		Locations []struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		}
		Path       []string `json:"path"`
		Extensions struct {
			Code         string `json:"code"`
			VariableName string `json:"variableName"`
		} `json:"extensions"`
	} `json:"errors"`
}

type Address struct {
	Address1      string `json:"address1"`
	Address2      string `json:"address2"`
	City          string `json:"city"`
	Company       string `json:"company"`
	Country       string `json:"country"`
	CountryCodeV2 string `json:"countryCodeV2"`
	FirstName     string `json:"firstName"`
	ID            string `json:"id"`
	LastName      string `json:"lastName"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	Zip           string `json:"zip"`
}

type Customer struct {
	Addresses          []Address `json:"addresses"`
	AmountSpent        MoneyV2   `json:"amountSpent"`
	CreatedAt          time.Time `json:"createdAt"`
	Currency           string    `json:"currency"`
	DefaultPhoneNumber struct {
		PhoneNumber string `json:"phoneNumber"`
	} `json:"defaultPhoneNumber"`
	DefaultEmailAddress struct {
		Email string `json:"emailAddress"`
	} `json:"defaultEmailAddress"`
	FirstName      string   `json:"firstName"`
	ID             string   `json:"id"`
	LastName       string   `json:"lastName"`
	NumberOfOrders string   `json:"numberOfOrders"`
	Phone          any      `json:"phone"`
	Tags           []string `json:"tags"`
	Ordered        string   `json:"ordered"`
}

func (c Customer) String() string {
	b, _ := json.MarshalIndent(c.Row(), "", "\t")
	return string(b)
}

func (c Customer) Row() any {
	var phone string
	if c.Phone != nil {
		phone = c.Phone.(string)
	} else {
		for _, a := range c.Addresses {
			if a.Phone != "" {
				phone = a.Phone
				break
			}
		}
	}
	numberOfOrders, _ := strconv.Atoi(c.NumberOfOrders)
	var city, state string
	for _, a := range c.Addresses {
		if city == "" && a.City != "" {
			city = a.City
		}
		if state == "" && a.Province != "" {
			state = a.Province
		}
		if city != "" && state != "" {
			break
		}
	}
	return map[string]any{
		"id":      c.ID,
		"name":    c.FirstName + " " + c.LastName,
		"tags":    c.Tags,
		"city":    city,
		"state":   state,
		"email":   c.DefaultEmailAddress.Email,
		"phone":   phone,
		"orders":  numberOfOrders,
		"ordered": c.Ordered,
		"spent":   c.AmountSpent.Decimal(),
	}
}

type MoneyBag struct {
	ShopMoney MoneyV2 `json:"shopMoney,omitempty"`
}

type MoneyV2 struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
}

func (m MoneyV2) Decimal() float64 {
	f, _ := strconv.ParseFloat(m.Amount, 64)
	return f
}

type Orders []Order

func (o Orders) Table() any {
	var arr []any
	for _, i := range o {
		arr = append(arr, i.Row())
	}
	return arr
}

func (o *Orders) UnmarshalJSON(b []byte) error {
	var g Graph[Order]
	if err := json.Unmarshal(b, &g); err != nil {
		return err
	}
	*o = g.Slice()
	return nil
}

type Order struct {
	CreatedAt             string    `json:"createdAt"`
	Customer              Customer  `json:"customer"`
	ID                    string    `json:"id"`
	TotalDiscountsSet     MoneyBag  `json:"totalDiscountsSet"`
	TotalPriceSet         MoneyBag  `json:"totalPriceSet"`
	TotalRefundedSet      MoneyBag  `json:"totalRefundedSet"`
	TotalShippingPriceSet MoneyBag  `json:"totalShippingPriceSet"`
	LineItems             LineItems `json:"lineItems"`
}

func (o Order) Row() any {
	return map[string]any{
		"id":        o.ID,
		"createdAt": o.CreatedAt,
		"customer":  o.Customer.FirstName + " " + o.Customer.LastName,
		"discounts": o.TotalDiscountsSet.ShopMoney.Decimal(),
		"price":     o.TotalPriceSet.ShopMoney.Decimal(),
		"refunded":  o.TotalRefundedSet.ShopMoney.Decimal(),
		"shipping":  o.TotalShippingPriceSet.ShopMoney.Decimal(),
		"items":     o.LineItems.Table(),
	}
}

type LineItems []LineItem

func (l *LineItems) Table() any {
	var arr []any
	for _, i := range *l {
		name := i.Variant.DisplayName
		if name == "" {
			name = i.Variant.Sku
		}
		if name == "" {
			name = i.Title
		}
		name = strings.TrimSuffix(name, " - Default Title")
		price, _ := strconv.ParseFloat(i.Variant.Price, 64)
		arr = append(arr, map[string]any{
			"id":       i.ID,
			"quantity": i.Quantity,
			"name":     name,
			"price":    price,
		})
	}
	return arr
}

func (l *LineItems) UnmarshalJSON(b []byte) error {
	var g Graph[LineItem]
	if err := json.Unmarshal(b, &g); err == nil {
		*l = g.Slice()
		return nil
	}
	var arr []LineItem
	err := json.Unmarshal(b, &arr)
	if err != nil {
		return err
	}
	*l = arr
	return nil
}

type LineItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
	Title    string `json:"title"`
	Variant  struct {
		GID         string `json:"id"`
		DisplayName string `json:"displayName"`
		Price       string `json:"price"`
		Sku         string `json:"sku"`
	} `json:"variant"`
}

type createArticleResponse struct {
	Data struct {
		ArticleCreate struct {
			Article struct {
				Author struct {
					Name string `json:"name"`
				} `json:"author"`
				Body   string `json:"body"`
				Handle string `json:"handle"`
				Id     string `json:"id"`
				Image  struct {
					AltText     string `json:"altText"`
					OriginalSrc string `json:"originalSrc"`
				} `json:"image"`
				Summary string   `json:"summary"`
				Tags    []string `json:"tags"`
				Title   string   `json:"title"`
			} `json:"article"`
			UserErrors []any `json:"userErrors"`
		} `json:"articleCreate"`
	} `json:"data"`
}

func PostArticle(
	token, store, blogId, title, author, handle, body, summary, imgSrc, imgAlt string,
	publishedAt time.Time,
	tags []string,
) ([]byte, error) {
	out, err := do(token, store, `mutation CreateArticle($article: ArticleCreateInput!) 
{ 
	articleCreate(article: $article) { 
		article { 
			id 
			title 
			author { name } 
			handle 
			body 
			summary 
			tags 
			image { altText originalSrc } 
		} 
		userErrors { code field message } 
	}
}`, map[string]any{
		"article": map[string]any{
			"blogId": blogId,
			"title":  title,
			"author": map[string]string{
				"name": author,
			},
			"handle":      handle,
			"body":        body,
			"summary":     summary,
			"isPublished": true,
			"publishDate": publishedAt,
			"tags":        tags,
			"image": map[string]string{
				"altText": imgAlt,
				"url":     imgSrc,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	var r createArticleResponse
	if err = json.Unmarshal(out, &r); err != nil {
		return nil, err
	}

	if len(r.Data.ArticleCreate.UserErrors) > 0 {
		b, _ := json.MarshalIndent(r.Data.ArticleCreate.UserErrors, "", "\t")
		return nil, errors.New("article create errors: " + string(b))
	}

	return out, nil
}

func GetOrders(token, store string, from, to time.Time) ([]Order, error) {
	q := `query GetOrders($first: Int!, $query: String!) 
{
	orders(first: $first, query: $query) {
		edges { 
			node { 
				id
				createdAt
				lineItems(first: 250) {	
					edges { 
						node {
							id
							quantity 
							title
							originalTotalSet { shopMoney { amount currencyCode } }
							variant { id displayName price sku } 
						} 
					}
				}
				totalPriceSet { shopMoney { amount currencyCode } }
				totalDiscountsSet { shopMoney { amount currencyCode } }
				totalRefundedSet { shopMoney { amount currencyCode } }
				totalShippingPriceSet { shopMoney { amount currencyCode } }
				customer {
					  id
					  firstName
					  lastName
					  createdAt
					  updatedAt
					  numberOfOrders
					  state
					  amountSpent { amount currencyCode }
					  defaultPhoneNumber { phoneNumber marketingState }
					  defaultEmailAddress { emailAddress marketingState }
					  verifiedEmail
					  tags
					  addresses(first: 100) {
						id
						firstName
						lastName
						address1
						city
						province
						country
						zip
						phone
						name
						countryCodeV2
					  }
					}
			} 
		}
  	}
}`
	v := map[string]any{
		"first": 250,
		"query": fmt.Sprintf(
			"created_at:>='%s' AND created_at:<='%s'",
			from.Format(time.RFC3339),
			to.Format(time.RFC3339),
		),
	}

	out, err := do(token, store, q, v)
	if err != nil {
		return nil, err
	}

	var r struct {
		Data struct {
			Orders Graph[Order] `json:"orders"`
		} `json:"data"`
	}
	if err = json.Unmarshal(out, &r); err != nil {
		return nil, err
	}

	return r.Data.Orders.Slice(), nil
}

func do(token, store, query string, variables any) ([]byte, error) {

	body, _ := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})

	out, err := https.PostJSON(
		fmt.Sprintf("https://%s.myshopify.com/admin/api/2026-01/graphql.json", store),
		body,
		map[string]string{"X-Shopify-Access-Token": token},
	)

	if err != nil {
		return nil, err
	}

	var r ErrorResponse
	_ = json.Unmarshal(out, &r)

	if len(r.Errors) > 0 {
		var messages []string
		for _, e := range r.Errors {
			messages = append(messages, e.Message)
		}
		return nil, errors.New(strings.Join(messages, "\n"))
	}

	return out, nil
}
