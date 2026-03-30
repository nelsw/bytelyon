package service

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/logger"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestSpinArticle(t *testing.T) {

	t.Setenv("MODE", "release")
	logger.Init()
	assert.NoError(t, godotenv.Load("../../.env"))
	assert.NoError(t, SpinArticle(
		ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"),
		ulid.MustParse("01KMYVPNZCCKC56FHGWG0WCE62"),
		ulid.MustParse("01KMYVQ2ACQQBR06TP4MASTXGV"),
	))
}
