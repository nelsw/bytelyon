package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AuthResponse events.APIGatewayV2CustomAuthorizerSimpleResponse
type Response events.APIGatewayV2HTTPResponse
type Request events.APIGatewayV2HTTPRequest

func (r Request) UserID() ulid.ULID {
	if _, ok := r.RequestContext.Authorizer.Lambda["userId"]; !ok {
		return ulid.Zero
	}
	id, _ := ulid.Parse(r.RequestContext.Authorizer.Lambda["userId"].(string))
	return id
}

func (r Request) BotType() model.BotType {
	bt := model.BotType(r.Query("type"))
	if err := bt.Validate(); err != nil {
		log.Warn().Err(err).Msg("invalid bot")
		return ""
	}
	return bt
}

func (r Request) Target() string {

	if r.BotType() != model.SitemapBotType {
		return r.Query("target")
	}

	b, err := base64.StdEncoding.DecodeString(r.Query("target"))
	if err != nil {
		log.Warn().Err(err).Msg("invalid target")
	}
	return string(b)
}

func (r Request) ID() ulid.ULID {
	if r.Query("id") == "" {
		return ulid.Zero
	}
	id, err := ulid.Parse(r.Query("id"))
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse id")
		return ulid.Zero
	}
	return id
}

func (r Request) Authorization() string { return r.Headers["authorization"] }
func (r Request) IsPreflight() bool     { return r.Method() == http.MethodOptions }
func (r Request) Query(k string) string { return r.QueryStringParameters[k] }
func (r Request) Method() string        { return r.RequestContext.HTTP.Method }

func (r Request) BAD(a any) Response     { return r.Response(http.StatusBadRequest, a) }
func (r Request) ERR(err error) Response { return r.Response(http.StatusInternalServerError, err) }
func (r Request) EX() Response           { return r.Response(http.StatusInternalServerError) }
func (r Request) NC() Response           { return r.Response(http.StatusNoContent) }
func (r Request) NI() Response           { return r.Response(http.StatusNotImplemented) }
func (r Request) OK(a any) Response      { return r.Response(http.StatusOK, a) }

func (r Request) Response(code int, a ...any) Response {

	log.Log().
		Dict("response", zerolog.Dict().
			Int("code", code).
			Bool("body", len(a) > 0)).
		Msg("response")

	var body string
	if len(a) == 0 {
		body = `{}`
	} else if err, ok := a[0].(error); ok {
		body = `{"message":"` + err.Error() + `"}`
	} else {
		var b []byte
		if b, err = json.MarshalIndent(a[0], "", "\t"); err != nil {
			body = `{"message":"` + err.Error() + `"}`
		} else {
			body = string(b)
		}
	}

	return Response{
		StatusCode: code,
		Body:       body,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "authorization, content-type,",
			"Access-Control-Allow-Methods": "*",
			"Content-BotType":              "application/json",
		},
	}
}

func (r Request) AuthOK(userID ulid.ULID, tkn string) AuthResponse {
	return r.AuthResponse(true, userID.String(), tkn)
}

func (r Request) AuthErr(err error) AuthResponse {
	return r.AuthResponse(false, err.Error())
}

func (r Request) AuthResponse(ok bool, s ...string) AuthResponse {

	log.Debug().Msgf("AuthResponse: %v:%s", ok, s)

	ctx := make(map[string]any)
	if !ok {
		ctx["error"] = s[0]
	} else {
		ctx["userId"] = s[0]
		ctx["token"] = s[1]
	}

	log.Log().
		Object("request", r).
		Dict("response", zerolog.Dict().
			Bool("isAuthorized", ok).
			Any("context", ctx)).
		Send()

	return AuthResponse{ok, ctx}
}

func (r Request) Log() { log.Log().EmbedObject(r).Msg("request") }
func (r Request) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("ip", r.RequestContext.HTTP.SourceIP).
		Str("method", r.Method()).
		Str("authorization", r.Authorization()).
		Str("path", r.RawPath).
		Any("query", r.QueryStringParameters).
		Str("body", r.Body)
}
