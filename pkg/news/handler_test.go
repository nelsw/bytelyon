package news

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Get_Headlines(t *testing.T) {
	logs.Init("debug")
	req := api.Request{
		QueryStringParameters: map[string]string{
			"topic": "ev fire",
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

func TestHandler_Get_Article(t *testing.T) {
	logs.Init("debug")
	req := api.Request{
		QueryStringParameters: map[string]string{
			"topic": "ev fire",
			"url":   "https://www.usfa.fema.gov/blog/emergency-response-to-electric-vehicle-incidents",
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
