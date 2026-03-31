package client

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestAccessToken(t *testing.T) {
	assert.NoError(t, godotenv.Load("../../.env"))
	tkn, err := accessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, tkn)
}
