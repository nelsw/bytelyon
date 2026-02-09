package news

type RSS struct {
	Channel struct {
		Items []*struct {
			URL         string `xml:"link"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Source      string `xml:"source"`
			Time        *Time  `xml:"pubDate"`
			NewsSource  string `xml:"News_Source"`
		} `xml:"item"`
	} `xml:"channel"`
}
