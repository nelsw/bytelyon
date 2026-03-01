package test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/client/ses"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func Test_Client_SES(t *testing.T) {
	ctx := context.Background()
	c := util.Must(client.New())
	to := "kowalski7012@gmail.com"

	assert.NoError(t, client.SendEmailConfirmation(ctx, c, to, uuid.NewString()))
	assert.NoError(t, client.SendPasswordReset(ctx, c, to, uuid.NewString()))
}
