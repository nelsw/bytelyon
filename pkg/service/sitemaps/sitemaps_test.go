package sitemaps

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {

	logs.Init()

	m := New("https://li-fire.com", 5)
	assert.NotEmpty(t, m)

	t.Log("urls: ", m.URLs.Len())
}
