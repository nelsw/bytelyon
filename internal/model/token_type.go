package model

type TokenType string

const (
	ResetPasswordTokenType TokenType = "reset"
	ConfirmEmailTokenType  TokenType = "confirm"
)
