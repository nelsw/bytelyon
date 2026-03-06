package main

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/rs/zerolog/log"
)

type Context map[string]any

type Claims struct {
	*model.User `json:"data"`
	jwt.RegisteredClaims
}

func NewClaims(u *model.User) *Claims {
	return &Claims{
		User: u,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ByteLyon API",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ID:        uuid.NewString(),
		},
	}
}

var (
	regex  = regexp.MustCompile("^(Bearer|Basic) (\\S+)$")
	secret = []byte(os.Getenv("JWT_SECRET"))
)

func Handler(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {

	log.Info().Any("Auth Request", req).Send()
	if req.RequestContext.HTTP.Method == http.MethodOptions {
		return response(true, Context{})
	}

	// if we're here, then we must have an authorization header, but the value has not been validated
	parts := strings.Split(req.Headers["authorization"], " ")

	if regex.MatchString(req.Headers["authorization"]) == false {
		return response(false, Context{"message": "invalid authorization token; must be 'Bearer <token>' or 'Basic <token>'"})
	}

	if parts[0] == "Bearer" {
		tkn, err := jwt.ParseWithClaims(parts[1], &Claims{}, func(*jwt.Token) (any, error) { return secret, nil })
		if err != nil {
			return response(false, Context{"message": err.Error()})
		}
		return response(true, Context{"user": tkn.Claims.(*Claims).User})
	}

	//a, err := model.NewBasicAuth(parts[1])
	//if err != nil {
	//	return response(false, Context{"message": err.Error()})
	//}

	var user *model.User

	//if user, err = a.Authenticate(); err != nil {
	//	return response(false, Context{"message": err.Error()})
	//}
	tkn, err := jwt.NewWithClaims(jwt.SigningMethodHS256, NewClaims(user)).SignedString(secret)
	if err != nil {
		return response(false, Context{"message": err.Error()})
	}

	return response(true, Context{"user": user, "token": tkn})
}

func response(ok bool, ctx Context) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {

	log.Info().
		Bool("isAuthorized", ok).
		Any("context", ctx).
		Msg("Auth Response")

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: ok,
		Context:      ctx,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
