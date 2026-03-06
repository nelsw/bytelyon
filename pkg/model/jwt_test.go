package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

type MyCustomClaims struct {
	ULID string `json:"ulid"`
	jwt.RegisteredClaims
}

func TestNewJWTv(t *testing.T) {
	// 2. Generate a new ULID
	newULID := ulid.Make().String()
	fmt.Println("Generated ULID:", newULID)

	// 3. Create claims
	claims := MyCustomClaims{
		newULID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 4. Create and Sign Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := []byte("your-secret-key")
	tokenString, _ := token.SignedString(secretKey)
	fmt.Println("JWT:", tokenString)

	// 5. Parse and Validate Token
	parsedToken, _ := jwt.ParseWithClaims(
		tokenString,
		&MyCustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

	// 6. Extract ULID
	if claims, ok := parsedToken.Claims.(*MyCustomClaims); ok && parsedToken.Valid {
		fmt.Println("Extracted ULID:", claims.ULID)
	}
}

func TestNewJWT(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-shared-key-32-chars-long!")

	//tkn, err := NewJWT(NewULID())
	//assert.NoError(t, err)
	//assert.NotEmpty(t, tkn)
	//t.Log(tkn)
	//
	//jwt, err := ParseJWT(tkn)
	//assert.NoError(t, err)
	//assert.NotEmpty(t, jwt)
	//t.Log(jwt)

	secret := []byte("070bb74c675267dc15a1f9466b115e57348326a30511d748712835745c5b64a8")
	user := NewUser()
	tkn, err := jwt.
		NewWithClaims(jwt.SigningMethodHS256, NewClaims(user)).
		SignedString(secret)

	fmt.Println(tkn, err)

	token, rerr := jwt.ParseWithClaims(tkn, &Claims{}, func(*jwt.Token) (any, error) { return secret, nil })
	fmt.Println(tkn, err)
	assert.NoError(t, rerr)
	txt, _ := user.ID.MarshalText()
	assert.Equal(t, txt, token.Claims.(*Claims).User)

	//jti := []byte(`450f7096-c2b6-4544-9eed-bd3daef049e8`)
	out := make([]byte, ulid.EncodedSize)
	err = user.ID.MarshalTextTo(out)
	assert.NoError(t, err)
	fmt.Println(ulid.MustParse(string(out)))
	fmt.Println(err)

	temp := ulid.Zero
	e := temp.UnmarshalText(out)
	fmt.Println(e, temp)
	fmt.Println(ulid.MustParse(string(out)))
}
