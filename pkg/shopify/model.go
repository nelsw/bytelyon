package shopify

import (
	"time"
)

type AdditionalFee struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Price MoneyBag `json:"price"`
}

type Address struct {
	Address1      string `json:"address1"`
	Address2      string `json:"address2"`
	City          string `json:"city"`
	Company       string `json:"company"`
	Country       string `json:"country"`
	CountryCodeV2 string `json:"country_code_v2"`
	CountryCode   string `json:"country_code"`
	CountryName   string `json:"country_name"`
	CustomerID    int64  `json:"customer_id"`
	Default       bool   `json:"default"`
	FirstName     string `json:"first_name"`
	ID            int    `json:"id"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	ProvinceCode  string `json:"province_code"`
	Zip           string `json:"zip"`
}

type AppliedDiscount struct {
	Amount      string  `json:"amount"`
	Description string  `json:"description"`
	Title       string  `json:"title"`
	Value       float64 `json:"value"`
	ValueType   string  `json:"valueType"`
}

type Customer struct {
	Addresses      []Address  `json:"addresses"`
	AmountSpent    MoneyV2    `json:"amount_spent"`
	CreatedAt      time.Time  `json:"created_at"`
	Currency       string     `json:"currency"`
	DefaultAddress Address    `json:"default_address"`
	DisplayName    string     `json:"display_name"`
	Email          string     `json:"email"`
	FirstName      string     `json:"first_name"`
	ID             int64      `json:"id"`
	LastName       string     `json:"last_name"`
	Locale         string     `json:"locale"`
	Metafields     Metafields `json:"metafields"`
	Note           string     `json:"note"`
	NumberOfOrders string     `json:"total_orders"`
	Phone          any        `json:"phone"`
	State          string     `json:"state"`
	Tags           []string   `json:"tags"`
	TaxExempt      bool       `json:"tax_exempt"`
	TaxExemptions  []any      `json:"tax_exemptions"`
	UpdatedAt      time.Time  `json:"updated_at"`
	VerifiedEmail  bool       `json:"verified_email"`
}

type MoneyBag struct {
	PresentmentMoney MoneyV2 `json:"presentmentMoney,omitempty"`
	ShopMoney        MoneyV2 `json:"shopMoney,omitempty"`
}

type MoneyV2 struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
}

func (r *MoneyBag) GraphQL() string { return `{ amount currencyCode }` }

type Metafield struct {
	JsonValue        bool      `json:"jsonValue"`
	Key              string    `json:"key"`
	LegacyResourceId string    `json:"legacyResourceId"`
	Namespace        string    `json:"namespace"`
	OwnerType        string    `json:"ownerType"`
	Type             string    `json:"type"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Value            string    `json:"value"`
	Id               string    `json:"id"`
	Description      any       `json:"description"`
	CompareDigest    string    `json:"compareDigest"`
	CreatedAt        time.Time `json:"createdAt"`
}

type Metafields struct {
	Nodes    []Metafield `json:"nodes"`
	PageInfo PageInfo    `json:"pageInfo"`
}

type Node struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Order struct {
	ID                    string   `json:"id"`
	CreatedAt             string   `json:"createdAt"`
	TotalPriceSet         MoneyBag `json:"totalPriceSet"`
	TotalDiscountsSet     MoneyBag `json:"totalDiscountsSet"`
	TotalRefundedSet      MoneyBag `json:"totalRefundedSet"`
	TotalShippingPriceSet MoneyBag `json:"totalShippingPriceSet"`
}

type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

type PaymentTerms struct {
	DueInDays        int    `json:"dueInDays"`
	Id               string `json:"id"`
	Overdue          bool   `json:"overdue"`
	PaymentTermsName string `json:"paymentTermsName"`
	PaymentTermsType string `json:"paymentTermsType"`
	TranslatedName   string `json:"translatedName"`
	PaymentSchedules struct {
		Nodes []struct {
			CompletedAt any       `json:"completedAt"`
			DueAt       time.Time `json:"dueAt"`
			Id          string    `json:"id"`
			IssuedAt    time.Time `json:"issuedAt"`
			Amount      struct {
				Amount       string `json:"amount"`
				CurrencyCode string `json:"currencyCode"`
			} `json:"amount"`
		} `json:"nodes"`
	} `json:"paymentSchedules" graphql:"paymentSchedules(first: 100)"`
}

type Variant struct {
	GID               string    `json:"id"`
	UpdatedAt         time.Time `json:"updatedAt"`
	DisplayName       string    `json:"displayName"`
	Price             string    `json:"price"`
	Sku               string    `json:"sku"`
	InventoryQuantity int       `json:"inventoryQuantity"`
}

type Weight struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}
