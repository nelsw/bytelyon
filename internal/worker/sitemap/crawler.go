package sitemap

type Crawler interface {
	Crawl(string) []string
}
