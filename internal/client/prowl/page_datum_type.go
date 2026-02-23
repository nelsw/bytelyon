package prowl

type DatumType string

const (
	SponsoredDatumType           DatumType = "sponsored"
	OrganicDatumType                       = "organic"
	VideoDatumType                         = "video"
	ForumDatumType                         = "forum"
	ArticleDatumType                       = "article"
	PopularProductsDatumType               = "popular_products"
	MoreProductsDatumType                  = "more_products"
	PeopleAlsoAskDatumType                 = "people_also_ask"
	PeopleAlsoSearchForDatumType           = "people_also_search_for"
)
