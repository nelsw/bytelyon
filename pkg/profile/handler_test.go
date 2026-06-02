package profile

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nelsw/bytelyon/pkg/anthropic"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/shopify"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/stretchr/testify/assert"
)

func Test_HandlePost(t *testing.T) {
	logs.Init("debug")

	req := api.HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				Lambda: map[string]any{
					"userId": "01KM010XK0HY8HWWFPJTZGRF0F",
				},
			},
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: http.MethodPut,
			},
		},
		Body: string(json.Of(&Model{
			Anthropic: anthropic.Credentials{},
			Img:       "https://bytelyon-public.s3.amazonaws.com/carl-headshot.png",
			Shopify:   shopify.Credentials{},
			Verified:  true,
		})),
	}

	res := Handler(req)
	res.Log()
	assert.NotEmpty(t, res.Body)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}
