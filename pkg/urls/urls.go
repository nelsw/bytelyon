package urls

import (
	"net/url"
	"path"
	"strings"
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

// PR returns the protocol-relative form of a URL; aka PRL (protocol-relative link).
// TLDR: this is just a fancy ƒ that trims https:// from a URL.
func PR(url string) string {
	// remove insecure protocol
	url = strings.TrimPrefix(Clean(url), "http://")
	// remove secure protocol
	return strings.TrimPrefix(url, "https://")
}

// Host returns the host name from a URL in lowercase.
// Unlinke url.Parse, this ƒ does not require a protocol to determine a hostname.
func Host(url string) string {
	// remove path
	url = strings.Split(PR(url), "/")[0]
	// remove query
	url = strings.Split(url, "?")[0]
	// remove fragment
	url = strings.Split(url, "#")[0]
	// remove port
	url = strings.Split(url, ":")[0]
	// lowercase
	return strings.ToLower(url)
}

// Domain returns the domain name from a URL in lowercase.
// Unlinke url.Parse, this ƒ does not require a protocol to determine a hostname.
func Domain(url string) string {

	url = Host(url)

	// remove subdomains
	for strings.Count(url, ".") > 1 {
		ss := strings.Split(url, ".")
		url = ss[len(ss)-2] + "." + ss[len(ss)-1]
	}

	return url
}

// HasFileExtension checks if the given URL path includes a file extension and returns true if it does, otherwise false.
func HasFileExtension(rawUrl string) bool {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return false
	}
	ext := path.Ext(u.Path) // Returns ".jpg", ".pdf", etc.
	return ext != ""
}
