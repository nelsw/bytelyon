package email

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/nelsw/bytelyon/pkg/aws"
	"github.com/nelsw/bytelyon/pkg/util/ptr"
	"github.com/rs/zerolog/log"
)

func SendEmailVerification(to, tkn string) error {
	html := strings.ReplaceAll(template, "{{href}}", os.Getenv("HOST")+"?tkn="+tkn)
	html = strings.ReplaceAll(html, "{{text}}", "Confirm Email")
	return sendEmail(to, `🦁 Confirm Email`, html)
}

func SendPasswordResetLink(to, tkn string) error {
	html := strings.ReplaceAll(template, "{{href}}", os.Getenv("HOST")+"?tkn="+tkn)
	html = strings.ReplaceAll(html, "{{text}}", "Reset Password")
	return sendEmail(to, `🦁 Reset Password`, html)
}

func sendEmail(to, subject, html string) (err error) {

	l := log.With().
		Str("ƒ", "sendEmail").
		Str("subject", subject).
		Str("to", to).
		Logger()

	l.Debug().Send()

	_, err = aws.SES.SendEmail(context.Background(), &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: ptr.Of("UTF-8"),
					Data:    &html,
				},
			},
			Subject: &types.Content{
				Charset: ptr.Of("UTF-8"),
				Data:    &subject,
			},
		},
		ReplyToAddresses: []string{"ByteLyon <no-reply@ByteLyon.com>"},
		Source:           ptr.Of("ByteLyon <no-reply@ByteLyon.com>"),
	})

	log.Err(err).Send()

	return
}
