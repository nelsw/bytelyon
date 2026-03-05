package model

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	"github.com/rs/zerolog/log"
)

var tokenErr = errors.New("invalid JWT token (either expired or unprocessable")

func NewJWT(userID ulid.ULID) (tkn string, err error) {

	ƒ := func(userID ulid.ULID) jwt.Claims {
		return jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "https://ByteLyon.com",
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			Subject:   userID.String(),
		}
	}

	log.Trace().Msg("creating JWT token")
	if tkn, err = jwt.NewWithClaims(jwt.SigningMethodHS256, ƒ(userID)).SignedString(JwtKey()); err != nil {
		log.Err(err).Msg("error creating JWT token")
	} else {
		log.Debug().Msg("created JWT token")
	}
	return
}

func ParseUserID(s string) (ulid.ULID, error) {

	log.Trace().Msg("parsing user id from JWT")

	id := ulid.Zero

	t, err := jwt.ParseWithClaims(s, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) { return JwtKey(), nil })

	if err != nil {
		log.Err(err).Msg("failed to parse user id (JWT parse err)")
		return id, err
	}

	if !t.Valid {
		log.Warn().Msg("unable to parse user id (JWT token invalid)")
		return id, tokenErr
	}

	if s, err = t.Claims.GetSubject(); err != nil {
		log.Err(err).Msg("unable to parse user id (JWT token subject err)")
		return id, err
	}

	if id, err = uuid.Parse(s); err != nil {
		log.Err(err).Msg("unable to parse user id (UUID parse err)")
		return id, err
	}

	log.Debug().Stringer("ID", id).Msg("parsed user id")

	return id, nil
}
