package model

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Token struct {
	Model
	Type   TokenType `json:"type" dynamodbav:"Type,string"`
	Expiry time.Time `json:"expiry" dynamodbav:"Expiry,number"`
}

func NewResetPasswordToken(userID ulid.ULID) *Token {
	return &Token{
		Model{UserID: userID},
		ResetPasswordTokenType,
		time.Now().Add(15 * time.Minute),
	}
}

func NewConfirmEmailToken(userID ulid.ULID) *Token {
	return &Token{
		Model{UserID: userID},
		ConfirmEmailTokenType,
		time.Now().Add(time.Hour),
	}
}
