package client

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestAccessToken(t *testing.T) {
	godotenv.Load()
	tkn, err := AccessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, tkn)
	t.Log(tkn)
}
