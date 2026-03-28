package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type ClaudeRequest struct {
	Model       string                `json:"model"`
	MaxTokens   int                   `json:"max_tokens"`
	Temperature int                   `json:"temperature"`
	Messages    []ClaudeMessage       `json:"messages"`
	Thinking    ClaudeRequestThinking `json:"thinking"`
}

type ClaudeMessage struct {
	Role    string                 `json:"role"`
	Content []ClaudeMessageContent `json:"content"`
}

type ClaudeMessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeRequestThinking struct {
	Type string `json:"type"`
}

type ClaudeResponse struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []ClaudeMessageContent
}

func NewSimpleClaudeRequest(text string) *ClaudeRequest {
	return &ClaudeRequest{
		Model:       "claude-sonnet-4-6",
		MaxTokens:   1000,
		Temperature: 1,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: []ClaudeMessageContent{{Type: "text", Text: text}},
			},
		},
		Thinking: ClaudeRequestThinking{Type: "disabled"},
	}
}

func DoClaudeRequest(r *ClaudeRequest) (string, error) {

	b, _ := json.Marshal(r)

	req, err := http.NewRequest(http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	req.Header.Set("anthropic-version", "2023-06-01")

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, _ = io.ReadAll(res.Body)

	var c ClaudeResponse
	if err = json.Unmarshal(b, &c); err != nil {
		return "", err
	}

	if len(c.Content) > 0 {
		return c.Content[0].Text, nil
	}

	return "", nil
}
