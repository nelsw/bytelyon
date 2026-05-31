package news

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

var googleRegex = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)

func decodeBingNewsLink(s string) string {
	s, _ = url.QueryUnescape(s)
	parts := strings.Split(s, "url=")
	if len(parts) < 2 {
		return s
	}
	return strings.Split(parts[1], "&amp;c=")[0]
}

func decodeGoogleURL(s string) string {

	matches := googleRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		log.Warn().Str("url", s).Msg("failed to match gstatic regex")
		return s
	}

	b, err := https.Get(s)
	if err != nil {
		log.Warn().Str("url", s).Msg("failed to get gstatic html")
		return s
	}

	var doc *html.Node
	if doc, err = html.Parse(bytes.NewReader(b)); err != nil {
		log.Warn().Str("url", s).Msg("failed to parse gstatic html")
		return s
	}

	var out string
	if out, err = decodeNode(doc, matches[1]); err != nil {
		log.Warn().Str("url", s).Msg("failed to decode gstatic node")
		return s
	}

	log.Trace().Str("out", out).Msg("decoded gstatic url")
	return out
}

func decodeNode(n *html.Node, encodedText string) (string, error) {
	if n.Type == html.ElementNode && n.Data == "c-wiz" {

		var sg, ts string
		if e := n.FirstChild; e != nil {
			for _, att := range e.Attr {
				switch att.Key {
				case "data-n-a-sg":
					sg = att.Val
				case "data-n-a-ts":
					ts = att.Val
				}
			}
		}
		return decodeParts(sg, ts, encodedText)
	}

	// continue traversing every sibling per child. give em noogies.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if u, e := decodeNode(c, encodedText); u != "" && e == nil {
			return u, nil
		}
	}
	return "", nil
}

func decodeParts(signature, timestamp, base64Str string) (string, error) {
	endpoint := "https://news.google.com/_/DotsSplashUi/data/batchexecute"
	payload := []any{
		"Fbv4je",
		fmt.Sprintf("[\"garturlreq\",[[\"X\",\"X\",[\"X\",\"X\"],null,null,1,1,\"US:en\",null,1,null,null,null,null,null,0,1],\"X\",\"X\",1,[1,1,1],1,1,null,0,0,null,0],\"%s\",%s,\"%s\"]", base64Str, timestamp, signature),
	}
	outer := [][]any{payload}
	bodyBytes, _ := json.Marshal([][][]any{outer})
	form := url.Values{}
	form.Set("f.req", url.QueryEscape(string(bodyBytes)))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString("f.req="+string(url.QueryEscape(string(bodyBytes)))))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	c := &http.Client{}
	var resp *http.Response
	if resp, err = c.Do(req); err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var b []byte
	if b, err = io.ReadAll(resp.Body); err != nil {
		return "", err
	}

	s := string(b)
	parts := strings.Split(s, "\n\n")
	if len(parts) < 2 {
		return "", errors.New("unexpected batchexecute response format")
	}

	payload = []any{}
	if err = json.Unmarshal([]byte(parts[1]), &payload); err != nil {
		return "", err
	} else if len(payload) == 0 {
		return "", errors.New("empty payload")
	}

	entry, ok := payload[0].([]any)
	if !ok || len(entry) < 3 {
		return "", errors.New("unexpected entry structure")
	}

	var inner []any
	if s, ok = entry[2].(string); !ok {
		return "", errors.New("missing inner json string")
	} else if err = json.Unmarshal([]byte(s), &inner); err != nil {
		return "", err
	} else if len(inner) < 2 {
		return "", errors.New("unexpected inner array")
	} else if s, ok = inner[1].(string); !ok {
		return "", errors.New("decoded url not string")
	}

	return s, nil
}

func chomp(in, a, z string) (string, string) {

	aIdx, zIdx := strings.Index(in, a), strings.Index(in, z)
	if aIdx == -1 || zIdx == -1 {
		return "", in
	}

	out := in[aIdx+len(a) : zIdx]

	in = in[:aIdx] + in[zIdx+len(z):]

	return out, in
}

func chomps(s, a, z string) (arr []string) {
	var res string
	for res, s = chomp(s, a, z); res != ""; res, s = chomp(s, a, z) {
		arr = append(arr, res)
	}
	return
}
