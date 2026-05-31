package api

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AuthRequest events.APIGatewayV2CustomAuthorizerV2Request
type AuthResponse events.APIGatewayV2CustomAuthorizerSimpleResponse

func Unauthorized(a any) AuthResponse {
	var msg string
	switch v := a.(type) {
	case string:
		msg = v
	case error:
		msg = v.Error()
	}
	return AuthResponse{
		Context: map[string]any{
			"message": msg,
		},
	}
}

func Authorized(tkn string, uid ulid.ULID) AuthResponse {
	return AuthResponse{
		IsAuthorized: true,
		Context: map[string]any{
			"userId": uid,
			"token":  tkn,
		},
	}
}

func (r AuthResponse) Log() { log.Log().EmbedObject(r).Msg("Auth Response") }

func (r AuthRequest) Authorization() (t, s string) {
	t, s, _ = strings.Cut(r.Headers["authorization"], " ")
	return
}

func (r AuthRequest) Log() { log.Log().EmbedObject(r).Msg("Auth Request") }

func (r AuthRequest) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("authorization", r.Headers["authorization"])
}

func (r AuthResponse) MarshalZerologObject(evt *zerolog.Event) {
	evt.Bool("isAuthorized", r.IsAuthorized)
	for k, v := range r.Context {
		evt.Any(k, v)
	}
}
