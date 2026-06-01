package api

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type HTTPResponse events.APIGatewayV2HTTPResponse
type HTTPRequest events.APIGatewayV2HTTPRequest

var headers = map[string]string{
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Headers": "authorization, content-type,",
	"Access-Control-Allow-Methods": "*",
	"Content-Type":                 "application/json",
}

func BadRequest(err error) HTTPResponse {
	return HTTPResponse{StatusCode: http.StatusBadRequest, Body: `{"message":"` + err.Error() + `"}`, Headers: headers}
}

func Forbidden() HTTPResponse {
	return HTTPResponse{StatusCode: http.StatusForbidden, Headers: headers}
}

func NoContent() HTTPResponse {
	return HTTPResponse{StatusCode: http.StatusNoContent, Headers: headers}
}

func NotImplemented() HTTPResponse {
	return HTTPResponse{StatusCode: http.StatusNotImplemented, Headers: headers}
}

func OK(a any) HTTPResponse {
	if b, err := json.Marshal(a); err != nil {
		return ServerError(err)
	} else {
		return HTTPResponse{StatusCode: http.StatusOK, Body: string(b), Headers: headers}
	}
}

func ServerError(err error) HTTPResponse {
	return HTTPResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       `{"message":"` + err.Error() + `"}`,
		Headers:    headers,
	}
}

func (r HTTPResponse) Log() { log.Log().EmbedObject(r).Msg("response") }
func (r HTTPResponse) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("status", r.StatusCode)
	if r.Body != "" {
		evt.Str("body", r.Body)
	}
}

func (r HTTPRequest) IsGuest() bool {
	s := r.UserID().String()
	return s != "01KMXGBJJE2GMCA1A9EXDGF4AJ" && s != "01KM010XK0HY8HWWFPJTZGRF0F"
}

func (r HTTPRequest) Log() { log.Log().EmbedObject(r).Msg("request") }

func (r HTTPRequest) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("method", r.RequestContext.HTTP.Method).
		Str("authorization", r.Headers["authorization"]).
		Str("path", r.RawPath).
		Any("query", r.QueryStringParameters).
		Str("body", r.Body)
}

func (r HTTPRequest) Query(k string) string { return r.QueryStringParameters[k] }

func (r HTTPRequest) UserID() ulid.ULID {
	if unk, ok := r.RequestContext.Authorizer.Lambda["userId"]; !ok {
		return ulid.Zero
	} else if s, k := unk.(string); k {
		return id.ParseULID(s)
	}
	return ulid.Zero
}
