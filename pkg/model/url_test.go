package model

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {

	s := "https://api.ByteLyon.com:8080/bots?type=news#latest"
	u, err := url.Parse(s)
	assert.NoError(t, err)
	assert.Equal(t, "api.ByteLyon.com", u.Hostname())
	t.Log(u.String())
	t.Log(u.Path)
	t.Log(u.Host)
	t.Log(u.Hostname())
	t.Log(u.EscapedPath())
	t.Log(u.Fragment)
	t.Log(u.Port())
	t.Log(u.Scheme)
	t.Log(u.User)
}
