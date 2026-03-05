package model

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type PageGraph map[PageSection][]*SectionDatum

func (p PageGraph) Item() map[string]types.AttributeValue {

	var item = make(map[string]types.AttributeValue)

	for section, data := range p {
		var items []types.AttributeValue
		for _, d := range data {
			items = append(items, &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"Position": &types.AttributeValueMemberN{Value: strconv.Itoa(d.Position)},
				"Title":    &types.AttributeValueMemberS{Value: d.Title},
				"Link":     &types.AttributeValueMemberS{Value: d.Link},
				"Source":   &types.AttributeValueMemberS{Value: d.Source},
				"Snippet":  &types.AttributeValueMemberS{Value: d.Snippet},
				"Price":    &types.AttributeValueMemberN{Value: strconv.FormatFloat(d.Price, 'f', -1, 64)},
			}})
		}
		item[string(section)] = &types.AttributeValueMemberL{Value: items}
	}

	return item
}

type PageSection string

const (
	SponsoredDatumType           PageSection = "sponsored"
	OrganicDatumType             PageSection = "organic"
	VideoDatumType               PageSection = "video"
	ForumDatumType               PageSection = "forum"
	ArticleDatumType             PageSection = "article"
	PopularProductsDatumType     PageSection = "popular_products"
	MoreProductsDatumType        PageSection = "more_products"
	PeopleAlsoAskDatumType       PageSection = "people_also_ask"
	PeopleAlsoSearchForDatumType PageSection = "people_also_search_for"
)

type SectionDatum struct {
	Position int     `json:"position"`
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Source   string  `json:"source"`
	Snippet  string  `json:"snippet"`
	Price    float64 `json:"price,omitempty"`
}
