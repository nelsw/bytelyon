package model

import (
	"errors"
	"net/mail"
	"unicode"

	"github.com/rs/zerolog/log"
)

var (
	errInvalidEmailAddress   = errors.New("invalid email address")
	errInvalidPasswordLen    = errors.New("password must contain at least 8 characters")
	errInvalidPasswordLower  = errors.New("password must contain at least one lowercase letter")
	errInvalidPasswordNumber = errors.New("password must contain at least one number")
	errInvalidPasswordSymbol = errors.New("password must contain at least one symbol")
	errInvalidPasswordUpper  = errors.New("password must contain at least one uppercase letter")
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Credentials) ValidateUsername() error {
	log.Trace().Str("email", c.Username).Msg("validating email address")
	if _, err := mail.ParseAddress(c.Username); err != nil {
		log.Warn().Err(err).Str("email", c.Username).Msg("invalid email address")
		return errors.Join(err, errInvalidEmailAddress)
	}
	log.Debug().Str("email", c.Username).Msg("valid email address")
	return nil
}

func (c *Credentials) ValidatePassword() error {

	if len(c.Password) < 8 {
		return errInvalidPasswordLen
	}

	var number, lower, upper, special bool
	for _, r := range c.Password {
		switch {
		case unicode.IsNumber(r):
			number = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			special = true
		case unicode.IsLetter(r) || r == ' ':
			lower = true
		}
	}

	if !lower {
		return errInvalidPasswordLower
	} else if !upper {
		return errInvalidPasswordUpper
	} else if !special {
		return errInvalidPasswordSymbol
	} else if !number {
		return errInvalidPasswordNumber
	}

	return nil
}

func NewCredentials(username string, password string) *Credentials {
	return &Credentials{username, password}
}
