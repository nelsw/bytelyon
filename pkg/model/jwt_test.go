package model

import (
	"fmt"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewJWT(t *testing.T) {
	t.Setenv("JWT_SECRET", "44493c0c-55bf-4bb2-b0f2-b68bee00df1e")
	tkn, err := NewJWT(NewID())
	assert.NoError(t, err)
	assert.NotEmpty(t, tkn)
	t.Log(tkn)

	jwt, err := ParseJWT(tkn)
	assert.NoError(t, err)
	assert.NotEmpty(t, jwt)
	t.Log(jwt)

	fmt.Println(ulid.MustParse("01KJZK4FDB5A1SKQ83RZNYC60A"))
}
