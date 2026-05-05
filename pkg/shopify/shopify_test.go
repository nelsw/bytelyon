package shopify

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestAccessToken(t *testing.T) {
	assert.NoError(t, godotenv.Load("../../.env"))
	tkn, err := accessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, tkn)
}

func TestGetOrderIds(t *testing.T) {
	assert.NoError(t, godotenv.Load("../../.env"))
	tkn, err := accessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, tkn)
	out, err := GetOrders(tkn, os.Getenv("SHOPIFY_STORE"), time.Now().Add(time.Hour*24*365*-1), time.Now())
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
	b, _ := json.MarshalIndent(out, "", "\t")
	fmt.Println(string(b))
}
