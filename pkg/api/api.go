package api

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AuthResponse events.APIGatewayV2CustomAuthorizerSimpleResponse
type Response events.APIGatewayV2HTTPResponse
type Request events.APIGatewayV2HTTPRequest

func Body[T any](req Request, in T) (err error) {
	if err = json.Unmarshal([]byte(req.Body), &in); err != nil {
		log.Err(err).Msgf("Failed to unmarshal body for [%t:%v]", in, in)
	}
	return
}

func (r Request) UserID() ulid.ULID {
	if id, ok := r.RequestContext.Authorizer.Lambda["userID"]; ok {
		return id.(ulid.ULID)
	}
	log.Warn().Msg("user id not found in request context?!")
	return ulid.Zero
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
			"Content-Type":                 "application/json",
		},
	}
}

func (r Request) AuthOK(userID ...ulid.ULID) AuthResponse { return r.AuthResponse(true, userID) }
func (r Request) AuthErr(a any) AuthResponse              { return r.AuthResponse(false, a) }
func (r Request) AuthResponse(ok bool, a ...any) AuthResponse {

	ctx := make(map[string]any)
	if len(a) > 0 {
		switch t := a[0].(type) {
		case ulid.ULID:
			ctx["userID"] = t
		case string:
			ctx["message"] = t
		case error:
			ctx["message"] = t.Error()
		}
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
		Str("query", r.RawQueryString).
		Bool("body", len(r.Body) > 0)
}
