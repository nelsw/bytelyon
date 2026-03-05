package model

type BotSitemap struct {
	Bot
	Target  string                      `json:"target"`
	Results map[string]BotSitemapResult `json:"results"`
}

type BotSitemapResult struct {
	Relative []string `json:"relative"`
	Remote   []string `json:"remote"`
}
