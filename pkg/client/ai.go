package client

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go" // imported as anthropic
	"github.com/rs/zerolog/log"
)

func Prompt(system string, message string) (string, error) {

	l := log.With().
		Str("system", system).
		Str("message", message).
		Logger()

	client := anthropic.NewClient()

	out, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
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

	return out.Content[0].Text, nil
}
