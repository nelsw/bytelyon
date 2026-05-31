package main

import (
	"encoding/base64"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/user"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func Handler(req api.AuthRequest) (res api.AuthResponse, err error) {

	req.Log()
	defer res.Log()

	switch typ, tkn := req.Authorization(); typ {
	case "Bearer":
		return handleBearerAuth(tkn), nil
	case "Basic":
		return handleBasicAuth(tkn), nil
	}

	return api.Unauthorized("invalid authorizer type; must be 'Bearer' or 'Basic'"), nil
}

func handleBasicAuth(tkn string) api.AuthResponse {

	b, err := base64.StdEncoding.DecodeString(tkn)
	if err != nil {
		return api.Unauthorized(err)
	}

	u, p, ok := strings.Cut(string(b), ":")
	if !ok {
		return api.Unauthorized("invalid basic token; must be base64 encoded '<email>:<password>'")
	}

	var usr *user.Model
	if usr, err = user.Login(u, p); err != nil {
		return api.Unauthorized(err)
	}

	tkn, err = jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    "ByteLyon API",
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ID:        usr.ID.String(),
	}).SignedString(jwtKey)

	if err != nil {
		return api.Unauthorized(err)
	}
	return api.Authorized(tkn, usr.ID)
}

func handleBearerAuth(t string) api.AuthResponse {

	tkn, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) { return jwtKey, nil })
	if err != nil {
		return api.Unauthorized(err)
	}

	uid := id.ParseULID(tkn.Claims.(*jwt.RegisteredClaims).ID)
	if !tkn.Valid || uid.IsZero() {
		return api.Unauthorized("invalid JWT token (either expired or unprocessable")
	}

	return api.Authorized(t, uid)
}

func main() { lambda.Start(Handler) }
