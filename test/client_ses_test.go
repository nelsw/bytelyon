package test

import (
	"context"
	"testing"

	"github.com/nelsw/bytelyon/internal/client/ses"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_Client_SES(t *testing.T) {
	ctx := context.Background()
	c := client.New(
		config.Get[string]("AWS_REGION"),
		config.Get[string]("AWS_ACCESS_KEY_ID"),
		config.Get[string]("AWS_SECRET_ACCESS_KEY"),
	)
	to := "kowalski7012@gmail.com"

	assert.NoError(t, client.SendEmailConfirmation(ctx, c, to))
	assert.NoError(t, client.SendPasswordReset(ctx, c, to))
}
