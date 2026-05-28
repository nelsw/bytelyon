package search

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Get_URLs(t *testing.T) {
	logs.Init("debug")
	req := api.Request{
		QueryStringParameters: map[string]string{
			"query": "ev fire blankets for sale",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				Lambda: map[string]any{
					"userId": "01KM010XK0HY8HWWFPJTZGRF0F",
				},
			},
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	}

	res := Handler(req)
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.NotEmpty(t, res.Body)
}

func TestHandler_Get_Snippet(t *testing.T) {
	logs.Init("debug")
	req := api.Request{
		QueryStringParameters: map[string]string{
			"domain": "firefibers.com",
			"url":    "https://firefibers.com",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				Lambda: map[string]any{
					"userId": "01KM010XK0HY8HWWFPJTZGRF0F",
				},
			},
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	}

	res := Handler(req)
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.NotEmpty(t, res.Body)
}
