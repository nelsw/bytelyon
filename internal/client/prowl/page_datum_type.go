package prowl

type DatumType string

const (
	SponsoredDatumType       DatumType = "sponsored"
	OrganicDatumType                   = "organic"
	VideoDatumType                     = "video"
	ForumDatumType                     = "forum"
	ArticleDatumType                   = "article"
	PopularProductsDatumType           = "popular_products"
	MoreProductsDatumType              = "more_products"
	RelatedQueryDatumType              = "related_query"
	AlsoAskedDatumType                 = "also_asked"
)
