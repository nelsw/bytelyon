package model

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type Claims struct {
	*User `json:"userID"`
	jwt.RegisteredClaims
}

func NewClaims(u *User) *Claims {
	return &Claims{
		User: u,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ByteLyon API",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ID:        string(u.ID.Bytes()),
		},
	}
}

var jwtErr = errors.New("invalid JWT token (either expired or unprocessable")
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func NewJWT(userID ulid.ULID) (tkn string, err error) {

	log.Trace().Msg("creating JWT token")
	if tkn, err = jwt.NewWithClaims(jwt.SigningMethodHS256, NewClaims(&User{ID: userID})).SignedString(jwtKey); err != nil {
		log.Err(err).Msg("error creating JWT token")
	} else {
		log.Debug().Msg("created JWT token")
	}
	return
}

func ParseJWT(s string) (ulid.ULID, error) {

	log.Trace().Msg("parsing user id from JWT")

	tkn, err := jwt.ParseWithClaims(s, &Claims{}, func(*jwt.Token) (any, error) { return jwtKey, nil })

	if err != nil {
		log.Err(err).Msg("failed to parse user id (JWT parse err)")
		return ulid.Zero, err
	}

	if !tkn.Valid {
		log.Warn().Msg("unable to parse user id (JWT token invalid)")
		return ulid.Zero, jwtErr
	}

	if s, err = tkn.Claims.GetSubject(); err != nil {
		log.Err(err).Msg("unable to parse user id (JWT token subject err)")
		return ulid.Zero, err
	}

	id := ulid.Zero
	if id, err = ulid.Parse(s); err != nil {
		log.Err(err).Msg("unable to parse user id (UUID parse err)")
		//return ulid.Zero, err
	}

	log.Debug().Stringer("ID", id).Msg("parsed jwt")

	return id, nil
}
