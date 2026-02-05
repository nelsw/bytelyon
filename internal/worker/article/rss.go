package article

type RSS struct {
	Channel struct {
		Items []*struct {
			URL    string `xml:"link"`
			Title  string `xml:"title"`
			Source string `xml:"source"`
			Time   *Time  `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}
