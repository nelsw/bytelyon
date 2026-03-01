package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

func SendEmail(ctx context.Context, c *ses.Client, to, subject, html string) error {

	l := log.With().
		Str("subject", subject).
		Str("to", to).
		Logger()

	l.Trace().Msgf("Sending email")

	_, err := c.SendEmail(ctx, &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: util.Ptr("UTF-8"),
					Data:    &html,
				},
			},
			Subject: &types.Content{
				Charset: util.Ptr("UTF-8"),
				Data:    &subject,
			},
		},
		ReplyToAddresses: []string{fmt.Sprintf("ByteLyon <no-reply@bytelyon.com>")},
		Source:           util.Ptr(fmt.Sprintf("ByteLyon <no-reply@bytelyon.com>")),
	})

	if err != nil {
		l.Err(err).Msg("failed to send email")
		return err
	}

	l.Debug().Msg("email sent")
	return nil
}
