package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AuthResponse events.APIGatewayV2CustomAuthorizerSimpleResponse
type Response events.APIGatewayV2HTTPResponse
type Request events.APIGatewayV2HTTPRequest

var headers = map[string]string{
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Headers": "authorization, content-type,",
	"Access-Control-Allow-Methods": "*",
	"Content-BotType":              "application/json",
}

func (r Request) Basic() (string, string, error) {

	_, tkn := r.Authorization()

	b, err := base64.StdEncoding.DecodeString(tkn)
	if err != nil {
		return "", "", err
	}

	u, p, ok := strings.Cut(string(b), ":")
	if !ok {
		err = errors.New("invalid basic token; must be base64 encoded '<email>:<password>'")
		return "", "", err
	}

	return u, p, nil
}

func (r Request) Authorization() (string, string) {
	t, s, _ := strings.Cut(r.Headers["authorization"], " ")
	return t, s
}

func (r Request) response(code int, a ...any) Response {

	var body string
	if len(a) == 0 {
		body = `{}`
	} else if err, ok := a[0].(error); ok {
		body = `{"message":"` + err.Error() + `"}`
	} else {
		var b []byte
		if b, err = json.Marshal(a[0]); err != nil {
			code = http.StatusInternalServerError
			body = `{"message":"` + err.Error() + `"}`
		} else {
			body = string(b)
		}
	}

	log.Info().Msgf("response: code=[%d] body=[%s]", code, body)

	return Response{
		StatusCode: code,
		Body:       body,
		Headers:    headers,
	}
}

func (r Request) Log()                   { log.Log().EmbedObject(r).Msg("request") }
func (r Request) Query(k string) string  { return r.QueryStringParameters[k] }
func (r Request) BAD(err error) Response { return r.response(http.StatusBadRequest, err) }
func (r Request) NC() Response           { return r.response(http.StatusNoContent) }
func (r Request) NI() Response           { return r.response(http.StatusNotImplemented) }
func (r Request) NOPE() Response         { return r.response(http.StatusForbidden) }
func (r Request) OK(a any) Response      { return r.response(http.StatusOK, a) }

func (r Request) Auth(a any) (res AuthResponse) {

	res.Context = make(map[string]any)
	if _, res.IsAuthorized = a.(string); !res.IsAuthorized {
		res.Context["message"] = a.(error).Error()
	} else {
		res.Context["token"] = a
	}

	log.Log().
		Object("request", r).
		Dict("response", new(zerolog.Event).CreateDict().
			Bool("isAuthorized", res.IsAuthorized).
			Any("context", res.Context)).
		Send()

	return
}

func (r Request) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("method", r.RequestContext.HTTP.Method).
		Str("authorization", r.Headers["authorization"]).
		Str("path", r.RawPath).
		Any("query", r.QueryStringParameters).
		Str("body", r.Body)
}

func (r Request) IsGuest() bool {
	return !regexp.
		MustCompile(`^(01KMXGBJJE2GMCA1A9EXDGF4AJ|01KM010XK0HY8HWWFPJTZGRF0F)$`).
		MatchString(r.UserID().String())
}

func (r Request) ID() ulid.ULID     { return r.id(r.Query("id")) }
func (r Request) UserID() ulid.ULID { return r.id(r.RequestContext.Authorizer.Lambda["userId"]) }
func (r Request) id(a any) (id ulid.ULID) {
	if s, ok := a.(string); ok {
		id, _ = ulid.Parse(s)
	}
	return id
}
