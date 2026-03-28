package client

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestDoClaudeRequest(t *testing.T) {
	godotenv.Load()
	text := "write a blog post summarizing this article: https://www.cbsnews.com/news/hurricane-milton-florida-electric-vehicles-ev-fire-risk/ as if you were a ev fire blanket sales person for company called FireFibers"
	req := NewSimpleClaudeRequest(text)
	txt, err := DoClaudeRequest(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, txt)
	t.Log(txt)
}
