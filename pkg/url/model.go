package url

import "strings"

type Model string

// Clean normalizes a URL string by trimming whitespace, converting to lowercase, and removing a trailing slash.
func (m Model) Clean() string {
	// trim whitespace jic
	s := strings.TrimSpace(m.String())
	// lowercase to normalize
	s = strings.ToLower(s)
	// remove trailing slash
	return strings.TrimSuffix(s, "/")
}

// Domain returns the domain name from a URL in lowercase.
// Unlinke url.Parse, this ƒ does not require a protocol to determine a hostname.
func (m Model) Domain() string {

	s := m.Host()

	// remove subdomains
	for strings.Count(s, ".") > 1 {
		ss := strings.Split(s, ".")
		s = ss[len(ss)-2] + "." + ss[len(ss)-1]
	}

	return s
}

// Host returns the host name from a URL in lowercase.
// Unlinke url.Parse, this ƒ does not require a protocol to determine a hostname.
func (m Model) Host() string {
	// remove path
	s := strings.Split(m.Clean(), "/")[0]
	// remove query
	s = strings.Split(s, "?")[0]
	// remove fragment
	s = strings.Split(s, "#")[0]
	// remove port
	s = strings.Split(s, ":")[0]
	// lowercase
	return strings.ToLower(s)
}

func (m Model) PR() string {
	// remove insecure protocol
	s := strings.TrimPrefix(m.Clean(), "http://")
	// remove secure protocol
	return strings.TrimPrefix(s, "https://")
}

func (m Model) String() string {
	return string(m)
}
