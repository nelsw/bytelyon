package article

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/rs/zerolog/log"
)

// FromRSS fetches articles from RSS feeds based on the provided topic.
func FromRSS(topic string, after time.Time, exclusions map[string]bool) (models []*Model) {
	models = append(models, FromGoogleNews(topic, after, exclusions)...)
	models = append(models, FromBingNews(topic, after, exclusions)...)
	models = append(models, FromBing(topic, after, exclusions)...)
	return
}

// FromGoogleNews fetches and parses news articles from Google News based on a given topic and returns them as an Articles slice.
// <item>
//
//	  <title>Here Are The GMC Sierra Discount, Lease And Finance Deals In October 2025 - GM Authority</title>
//		 <link>https://news.google.com/rss/articles/CBMirAFBVV95cUxQU...</link>
//		 <guid isPermaLink="false">CBMirAFBVV95cUx...</guid>
//		 <pubDate>Thu, 02 Oct 2025 07:00:00 GMT</pubDate>
//		 <description>
//		   <a href="https://news.google.com/rss/articles/CBMirAFBVV95cUxQU..." target="_blank">Here Are The GMC Sierra Discount, Lease And Finance Deals In October 2025</a>
//		   <font color="#6f6f6f">GM Authority</font>
//		 </description>
//		 <source url="https://gmauthority.com">GM Authority</source>
//
// </item>
func FromGoogleNews(topic string, after time.Time, exclusions map[string]bool) (articles []*Model) {

	q := strings.ReplaceAll(topic, " ", "+")
	u := fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q)

	l := log.With().
		Str("ƒ", "FromGoogleNews").
		Str("url", u).
		Str("topic", topic).
		Time("after", after).
		Any("exclusions", exclusions).
		Logger()

	l.Trace().Send()

	b, err := https.Get(u)
	if err != nil {
		l.Err(err).Send()
		return
	}

	for _, item := range chomps(string(b), "<item>", "</item>") {

		var title, link, pubDate, source string

		pubDate, item = chomp(item, "<pubDate>", "</pubDate>")
		t, _ := time.Parse(time.RFC1123, pubDate)
		if t.Before(after) {
			continue
		}

		link, item = chomp(item, "<link>", "</link>")
		link = decodeGoogleURL(link)

		title, item = chomp(item, "<title>", "</title>")
		if parts := strings.Split(title, " - "); len(parts) > 1 {
			title = parts[0]
			source = parts[1]
		} else {
			source, item = chomp(item, `">`, "</source>")
		}

		var exclusionCount int
		for exclusion := range exclusions {
			exclusionCount += strings.Count(link, exclusion)
			exclusionCount += strings.Count(title, exclusion)
			exclusionCount += strings.Count(source, exclusion)
		}
		if exclusionCount > 0 {
			l.Debug().Str("url", link).Int("exclusionCount", exclusionCount).Send()
			continue
		}

		articles = append(articles, &Model{
			ID:     model.NewULID(t),
			Image:  model.MakeImage("", ""),
			Source: source,
			Title:  title,
			URL:    link,
		})
	}

	l.Debug().Int("articles", len(articles)).Send()

	return
}

// FromBingNews fetches articles from bing news rss feed; example:
// <item>
//
//	  <title>Situation With Iran Is Reaching A Boiling Point</title>
//		 <link>http://www.bing.com/news/apiclick.aspx?ref=FexRss&aid=&tid=6a1296fa090c44d1b6ddcf31945d68a5&url=https%3a%2f%2fwibc.com%2f887629%2fsituation-with-iran-is-reaching-a-boiling-point%2f&c=14824072920173618629&mkt=en-us</link>
//		 <description>The discussion then shifts to the Middle East, where the situation with Iran is reaching a boiling point. Major Lyons is adamant that the US needs to take a firmer stance, saying “I believe that the ...</description>
//		 <pubDate>Fri, 22 May 2026 02:51:00 GMT</pubDate>
//		 <News:Source>WIBC 93.1 FM</News:Source>
//		 <News:Image>http://www.bing.com/th?id=ONUT.735NYY5Sp5uj8m5xwyRzuQ&pid=News</News:Image>
//		 <News:ImageSize>w={0}&h={1}&c=14</News:ImageSize>
//		 <News:ImageMaxWidth>724</News:ImageMaxWidth>
//		 <News:ImageMaxHeight>483</News:ImageMaxHeight>
//
// </item>
func FromBingNews(topic string, after time.Time, exclusions map[string]bool) (articles []*Model) {

	u := fmt.Sprintf(
		"https://www.bing.com/news/search?format=rss&q=%s",
		strings.ReplaceAll(topic, " ", "+"),
	)

	l := log.With().
		Str("ƒ", "FromBingNews").
		Str("url", u).
		Str("topic", topic).
		Time("after", after).
		Any("exclusions", exclusions).
		Logger()

	l.Trace().Send()

	b, err := https.Get(u)
	if err != nil {
		l.Err(err).Send()
		return
	}

	for _, item := range chomps(string(b), "<item>", "</item>") {

		var title, link, pubDate, description, source, image, imageMaxWidth, imageMaxHeight string

		pubDate, item = chomp(item, "<pubDate>", "</pubDate>")
		t, _ := time.Parse(time.RFC1123, pubDate)
		if t.Before(after) {
			continue
		}

		title, item = chomp(item, "<title>", "</title>")
		description, item = chomp(item, "<description>", "</description>")
		source, item = chomp(item, "<News:Source>", "</News:Source>")
		link, item = chomp(item, "<link>", "</link>")
		link, _ = url.QueryUnescape(link)

		link, _ = chomp(link, "url=", "&amp;c=")

		var exclusionCount int
		for exclusion := range exclusions {
			exclusionCount += strings.Count(link, exclusion)
			exclusionCount += strings.Count(title, exclusion)
			exclusionCount += strings.Count(description, exclusion)
			exclusionCount += strings.Count(source, exclusion)
		}
		if exclusionCount > 0 {
			l.Debug().Str("url", link).Int("exclusionCount", exclusionCount).Send()
			continue
		}

		image, item = chomp(item, "<News:Image>", "</News:Image>")
		imageMaxWidth, item = chomp(item, "<News:ImageMaxWidth>", "</News:ImageMaxWidth>")
		imageMaxHeight, item = chomp(item, "<News:ImageMaxHeight>", "</News:ImageMaxHeight>")
		imageUrl := fmt.Sprintf("%s&w=%s&h=%s", image, imageMaxWidth, imageMaxHeight)

		articles = append(articles, &Model{
			Description: description,
			ID:          model.NewULID(t),
			Image:       model.MakeImage(imageUrl, ""),
			Source:      source,
			Title:       title,
			URL:         link,
		})
	}

	l.Debug().Int("articles", len(articles)).Send()

	return
}

// FromBing retrieves a list of articles from Bing based on the provided topic.
// <item>
//
//	<title>Live Updates: Peace deal with Iran has been "largely negotiated" and ...</title>
//	<link>https://www.cbsnews.com/live-updates/iran-war-trump-us-peace-talks-strait-of-hormuz-control/</link>
//	<description>Following a call with several Middle Eastern leaders, President Trump said that a peace deal with Iran had been "largely negotiated."</description>
//	<pubDate>Sat, 23 May 2026 23:29:00 GMT</pubDate>
//
// </item>
func FromBing(topic string, after time.Time, exclusions map[string]bool) (models []*Model) {
	u := fmt.Sprintf(
		"https://www.bing.com/search?format=rss&q=%s",
		strings.ReplaceAll(topic, " ", "+"),
	)

	l := log.With().
		Str("ƒ", "FromBing").
		Str("url", u).
		Str("topic", topic).
		Time("after", after).
		Any("exclusions", exclusions).
		Logger()

	l.Trace().Send()

	b, err := https.Get(u)
	if err != nil {
		l.Err(err).Send()
		return
	}

	for _, item := range chomps(string(b), "<item>", "</item>") {
		fmt.Println(item)
		var title, link, pubDate, description string

		pubDate, item = chomp(item, "<pubDate>", "</pubDate>")
		t, _ := time.Parse(time.RFC1123, pubDate)
		if t.Before(after) {
			continue
		}

		title, item = chomp(item, "<title>", "</title>")
		description, item = chomp(item, "<description>", "</description>")
		link, item = chomp(item, "<link>", "</link>")
		link = urls.Clean(link)

		var exclusionCount int
		for exclusion := range exclusions {
			exclusionCount += strings.Count(link, exclusion)
			exclusionCount += strings.Count(title, exclusion)
			exclusionCount += strings.Count(description, exclusion)
		}
		if exclusionCount > 0 {
			l.Debug().Str("url", link).Int("exclusionCount", exclusionCount).Send()
			continue
		}

		models = append(models, &Model{
			Description: description,
			ID:          model.NewULID(t),
			Image:       model.MakeImage("", ""),
			Title:       title,
			URL:         link,
		})
	}

	l.Debug().Int("articles", len(models)).Send()

	return
}
