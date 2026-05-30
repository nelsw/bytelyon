package main

import (
	"errors"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/user"
	"github.com/rs/zerolog/log"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func Handler(r api.Request) (any, error) {
	r.Log()
	if typ, tkn := r.Authorization(); typ == "Bearer" {
		return handleBearerAuth(r, tkn)
	} else if typ == "Basic" {
		return handleBasicAuth(r)
	}
	return r.Auth(errors.New("invalid authorizer type; must be 'Bearer' or 'Basic'")), nil
}

func handleBasicAuth(r api.Request) (any, error) {

	var u *user.Model
	if e, p, err := r.Basic(); err != nil {
		return nil, err
	} else if u, err = user.Login(e, p); err != nil {
		return nil, err
	}

	tkn, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    "ByteLyon API",
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ID:        u.ID.String(),
	}).SignedString(jwtKey)

	if err != nil {
		return nil, err
	}

	// todo - is this necessary?
	if r.Query("action") == "login" {
		return r.OK(map[string]any{"token": tkn}), nil
	}

	return r.Auth(tkn), nil
}

func handleBearerAuth(r api.Request, t string) (api.AuthResponse, error) {

	tkn, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) { return jwtKey, nil })
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse jwt")
		return r.Auth(err), nil
	}

	if !tkn.Valid || id.ParseULID(tkn.Claims.(*jwt.RegisteredClaims).ID).IsZero() {
		err = errors.New("invalid JWT token (either expired or unprocessable")
		log.Warn().Err(err).Send()
		return r.Auth(err), nil
	}

	log.Debug().Msg("jwt valid")
	return r.Auth(t), nil
}

func main() { lambda.Start(Handler) }
