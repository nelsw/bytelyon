package model

type TokenType string

const (
	ResetPasswordTokenType TokenType = "ResetPasswordToken"
	ConfirmEmailTokenType  TokenType = "ConfirmEmailToken"
)
