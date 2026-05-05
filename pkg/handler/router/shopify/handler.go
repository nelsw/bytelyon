package shopify

import (
	"maps"
	"net/http"
	"slices"
	"time"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/shopify"
	"github.com/nelsw/bytelyon/pkg/store"
)

func Handler(r Request) Response {

	r.Log()

	if r.IsGuest() {
		return r.NOPE()
	}

	switch r.Method() {
	case http.MethodPost:
		return handlePost(r)
	case http.MethodGet:
		return handleGet(r)
	}

	return r.NI()
}

func handlePost(r Request) Response {
	return r.OK("test")
}

func handleGet(r Request) Response {

	orderDB, err := store.New[shopify.Order]("orders.json")
	if err != nil {
		return r.BAD(err)
	}
	orderDB.Close()

	var orders []any
	customers := map[string]any{}
	for _, order := range orderDB.All() {
		orders = append(orders, order.Row())
		c := order.Customer
		c.Ordered = order.CreatedAt
		if _, ok := customers[order.Customer.ID]; !ok {
			customers[c.ID] = c.Row()
			continue
		}
		a, _ := time.Parse(time.RFC3339, c.Ordered)
		b, _ := time.Parse(time.RFC3339, customers[c.ID].(map[string]any)["ordered"].(string))
		if a.After(b) {
			customers[c.ID] = c.Row()
		}
	}

	return r.OK(map[string]any{
		"customers": slices.Collect(maps.Values(customers)),
		"orders":    orders,
	})
}
