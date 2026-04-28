package client

import (
	"bytes"
	"context"

	"github.com/anthropics/anthropic-sdk-go" // imported as anthropic
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
)

var ai = anthropic.NewClient()

func Prompt(system string, message string, html ...bool) (string, error) {

	l := log.With().
		Str("system", system).
		Str("message", message).
		Logger()

	out, err := ai.Messages.New(context.Background(), anthropic.MessageNewParams{
		MaxTokens: 2048,
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(message))},
		System:    []anthropic.TextBlockParam{{Text: system}},
		Model:     anthropic.ModelClaudeOpus4_6,
	})

	if err != nil {
		l.Err(err).Msg("anthropic prompt failed")
		return "", err
	}

	l.Info().Any("response", out).Msg("anthropic prompt succeeded")

	txt := out.Content[0].Text
	if len(html) == 0 || !html[0] {
		return txt, nil
	}

	var buf bytes.Buffer
	if err = goldmark.Convert([]byte(txt), &buf); err != nil {
		log.Err(err).Msg("failed to convert article from md to html")
		return "", err
	}

	return buf.String(), nil
}

func SimpleUserMessage(ctx context.Context, client *anthropic.Client, system, message string) (string, error) {

	out, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 2048,
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(message))},
		System:    []anthropic.TextBlockParam{{Text: system}},
		Model:     anthropic.ModelClaudeOpus4_6,
	})

	if err == nil && len(out.Content) > 0 {
		return out.Content[0].Text, nil
	}

	return "", err
}
