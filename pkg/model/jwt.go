package model

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var jwtErr = errors.New("invalid JWT token (either expired or unprocessable")
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func NewJWT(userID ulid.ULID) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    "ByteLyon API",
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ID:        userID.String(),
	}).SignedString(jwtKey)
}

func ParseJWT(s string) (ulid.ULID, error) {

	log.Trace().Msg("parsing user id from JWT")

	tkn, err := jwt.ParseWithClaims(s, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) {
		return jwtKey, nil
	})

	if err != nil {
		log.Err(err).Msg("failed to parse user id (JWT parse err)")
		return ulid.Zero, err
	}

	if !tkn.Valid {
		log.Warn().Msg("unable to parse user id (JWT token invalid)")
		return ulid.Zero, jwtErr
	}

	var id ulid.ULID
	id, err = ulid.Parse(tkn.Claims.(*jwt.RegisteredClaims).ID)

	if err != nil {
		log.Err(err).Msg("unable to parse user id (JWT token invalid)")
		return ulid.Zero, jwtErr
	}

	if id == ulid.Zero {
		log.Warn().Msg("unable to parse user id (JWT token invalid)")
	} else {
		log.Debug().Msg("parsed user id from JWT")
	}

	return id, nil
}
