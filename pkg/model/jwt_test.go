package model

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

type MyCustomClaims struct {
	ULID string `json:"ulid"`
	jwt.RegisteredClaims
}

var secret = []byte("070bb74c675267dc15a1f9466b115e57348326a30511d748712835745c5b64a8")

func Test_JWT(t *testing.T) {
	config.Init()
	logger.Init()
	t.Setenv("JWT_SECRET", "070bb74c675267dc15a1f9466b115e57348326a30511d748712835745c5b64a8")

	expID := NewULID()
	expTkn, err := NewJWT(expID)
	assert.NoError(t, err)
	t.Log(expID)

	var actID ulid.ULID
	actID, err = ParseJWT(expTkn)
	assert.NoError(t, err)
	t.Log(actID)

	assert.Equal(t, expID, actID)
}
