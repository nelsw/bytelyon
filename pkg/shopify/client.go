package shopify

import (
	"encoding/json"
	"errors"
	"fmt"
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

type OrdersResponse struct {
	Data struct {
		Orders struct {
			Edges []struct {
				Node Order `json:"node"`
			} `json:"edges"`
		} `json:"orders"`
	} `json:"data"`
}

func (r OrdersResponse) Orders() []Order {
	var orders []Order
	for _, o := range r.Data.Orders.Edges {
		orders = append(orders, o.Node)
	}
	return orders
}

func createArticle(
	token, store, blogId, title, author, handle, body, summary, imgSrc, imgAlt string,
	publishedAt time.Time,
	tags []string,
) ([]byte, error) {
	return do(token, store, `mutation CreateArticle($article: ArticleCreateInput!) 
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
}

func getOrder(token, store, orderId string) ([]byte, error) {
	val := `query {
  order(id: "$orderId") {
    id
    name
	customer {
	  id
	  firstName
	  lastName
	  createdAt
	  updatedAt
	  numberOfOrders
	  state
	  amountSpent { amount, currencyCode }
	  defaultPhoneNumber { phoneNumber, marketingState }
	  defaultEmailAddress { emailAddress, marketingState }
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
    totalPriceSet {
      presentmentMoney {
        amount
      }
    }
    lineItems(first: 10) {
      nodes {
        id
        name
      }
    }
  }
}`
	return do(token, store, val, map[string]string{
		"orderId": orderId,
	})
}

func getOrders(token, store string, from, to time.Time) ([]Order, error) {
	q := `query Input($first: Int!, $query: String!) 
{
	orders(first: $first, query: $query) {
		edges { 
			node { 
				id
				createdAt
lineItems
				totalPriceSet {
					presentmentMoney { amount currencyCode }
					shopMoney { amount currencyCode }
				}
				totalDiscountsSet {
					presentmentMoney { amount currencyCode }
					shopMoney { amount currencyCode }
				}
				totalRefundedSet {
					presentmentMoney { amount currencyCode }
					shopMoney { amount currencyCode }
				}
				totalShippingPriceSet {
					presentmentMoney { amount currencyCode }
					shopMoney { amount currencyCode }
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

	var a any
	_ = json.Unmarshal(out, &a)
	b, _ := json.MarshalIndent(a, "", "\t")
	fmt.Println(string(b))

	var r OrdersResponse
	if err = json.Unmarshal(out, &r); err != nil {
		return nil, err
	}

	return r.Orders(), nil
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
