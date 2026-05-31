package urls

import (
	"regexp"
	"strings"
)

var (
	browserFunction = regexp.MustCompile(`^(mailto|tel|sms|fax|callto|geo|javascript|about):.*`)
)

// Clean normalizes a URL string by trimming whitespace, converting to lowercase, and removing a trailing slash.
func Clean(url string) string {
	// trim whitespace jic
	url = strings.TrimSpace(url)
	// lowercase to normalize
	url = strings.ToLower(url)
	// remove trailing slash
	return strings.TrimSuffix(url, "/")
}

// Domain returns the domain name from a URL in lowercase.
// Unlinke url.Parse, this ƒ does not require a protocol to determine a hostname.
func Domain(url string) string {

	// remove protocols
	url = strings.TrimPrefix(Clean(url), "http://")
	url = strings.TrimPrefix(url, "https://")

	// remove path
	url = strings.Split(url, "/")[0]

	// remove query
	url = strings.Split(url, "?")[0]

	// remove fragment
	url = strings.Split(url, "#")[0]

	// remove port
	url = strings.Split(url, ":")[0]

	// remove subdomains
	for strings.Count(url, ".") > 1 {
		ss := strings.Split(url, ".")
		url = ss[len(ss)-2] + "." + ss[len(ss)-1]
	}

	return url
}

func IsBrowserFunction(s string) bool {
	return browserFunction.MatchString(s)
}
