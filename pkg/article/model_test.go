package article

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestFromRSS(t *testing.T) {
	logs.Init("trace")
	topic := "situation in iran"
	after := time.Now().AddDate(0, 0, -2)

	out := FromBingNews(topic, after, nil)
	assert.NotEmpty(t, out)

	util.PrintlnPrettyJSON(out)
	fmt.Println(out[0].URL)
}

func TestChomps(t *testing.T) {
	s := `<item><title>Iran War: Latest Breaking News, Updates &amp; Analysis | Reuters</title><link>https://www.reuters.com/world/iran/</link><description>Real-time Reuters coverage of the Iran war: US-Israel strikes, Iranian retaliation, nuclear threats, oil market shocks, and regional war risks.</description><pubDate>Tue, 31 Mar 2026 00:33:00 GMT</pubDate></item>`
	arr := chomps(s+s, "<item>", "</item>")
	assert.NotEmpty(t, arr)
	util.PrintlnPrettyJSON(arr)
}

func TestChomp(t *testing.T) {
	in := `<title>Iran War: Latest Breaking News, Updates &amp; Analysis | Reuters</title><link>https://www.reuters.com/world/iran/</link><description>Real-time Reuters coverage of the Iran war: US-Israel strikes, Iranian retaliation, nuclear threats, oil market shocks, and regional war risks.</description><pubDate>Tue, 31 Mar 2026 00:33:00 GMT</pubDate>`
	t.Log(in)
	a := strings.Index(in, "<title>")
	t.Log(a)
	z := strings.Index(in, "</title>")
	t.Log(z)
	out := in[a+len("<title>") : z]
	t.Log(out)
	in = in[:a] + in[z+len("</title>"):]
	t.Log(in)
}
