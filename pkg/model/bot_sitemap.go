package model

type BotSitemap struct {
	Bot
	Results map[string]BotSitemapResult `json:"results"`
}

type BotSitemapResult struct {
	Relative []string `json:"relative"`
	Remote   []string `json:"remote"`
}
