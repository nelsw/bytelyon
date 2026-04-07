package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var isStu = regexp.MustCompile(`^(01KMXGBJJE2GMCA1A9EXDGF4AJ|01KM010XK0HY8HWWFPJTZGRF0F|01KM01JC9PS1R4X4FDJNFAR4AZ)$`)

type AuthResponse events.APIGatewayV2CustomAuthorizerSimpleResponse
type Response events.APIGatewayV2HTTPResponse
type Request events.APIGatewayV2HTTPRequest

func (r Request) Authorization() string  { return r.Headers["authorization"] }
func (r Request) IsPreflight() bool      { return r.Method() == http.MethodOptions }
func (r Request) Log()                   { log.Log().EmbedObject(r).Msg("request") }
func (r Request) Method() string         { return r.RequestContext.HTTP.Method }
func (r Request) Query(k string) string  { return r.QueryStringParameters[k] }
func (r Request) BAD(a any) Response     { return r.Response(http.StatusBadRequest, a) }
func (r Request) NOPE() Response         { return r.Response(http.StatusForbidden) }
func (r Request) ERR(err error) Response { return r.Response(http.StatusInternalServerError, err) }
func (r Request) EX() Response           { return r.Response(http.StatusInternalServerError) }
func (r Request) NC() Response           { return r.Response(http.StatusNoContent) }
func (r Request) NI() Response           { return r.Response(http.StatusNotImplemented) }
func (r Request) OK(a any) Response      { return r.Response(http.StatusOK, a) }

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
		Dict("response", new(zerolog.Event).CreateDict().
			Bool("isAuthorized", ok).
			Any("context", ctx)).
		Send()

	return AuthResponse{ok, ctx}
}

func (r Request) Response(code int, a ...any) Response {

	if len(a) == 0 {
		a = append(a, nil)
	}

	log.Log().
		Dict("response", new(zerolog.Event).CreateDict().
			Int("code", code).
			Any("body", a[0])).
		Msg("response")

	var body string
	if len(a) == 0 || a[0] == nil {
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

func (r Request) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("ip", r.RequestContext.HTTP.SourceIP).
		Str("method", r.Method()).
		Str("authorization", r.Authorization()).
		Str("path", r.RawPath).
		Any("query", r.QueryStringParameters).
		Str("body", r.Body)
}

func (r Request) BotType() model.BotType {
	if err := model.BotType(r.Query("type")).Validate(); err != nil {
		log.Warn().Err(err).Msg("invalid bot")
		return ""
	}
	return model.BotType(r.Query("type"))
}

func (r Request) Target() string {
	if r.BotType() != model.SitemapBotType {
		return r.Query("target")
	}
	return strings.ReplaceAll(r.Query("target"), " ", ".")
}

func (r Request) UserID() ulid.ULID {
	return r.id(r.RequestContext.Authorizer.Lambda["userId"])
}

func (r Request) BotID() ulid.ULID {
	return r.id(r.Query("botId"))
}

func (r Request) ID() ulid.ULID {
	return r.id(r.Query("id"))
}

func (r Request) IsStu() bool {
	return isStu.MatchString(r.UserID().String())
}

func (r Request) id(a any) ulid.ULID {

	s, ok := a.(string)
	if !ok {
		return ulid.Zero
	} else if s == "" {
		return ulid.Zero
	}

	id, err := ulid.Parse(s)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse id for " + s)
		return ulid.Zero
	}

	return id
}
