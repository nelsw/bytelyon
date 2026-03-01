package model

import (
	"errors"
	"net/mail"
	"unicode"

	"github.com/rs/zerolog/log"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Credentials) ValidateUsername() error {
	log.Trace().Str("email", c.Username).Msg("validating email address")
	if _, err := mail.ParseAddress(c.Username); err != nil {
		log.Warn().Err(err).Str("email", c.Username).Msg("invalid email address")
		return err
	}
	log.Debug().Str("email", c.Username).Msg("valid email address")
	return nil
}

func (c *Credentials) ValidatePassword() (err error) {

	if len(c.Password) < 8 {
		err = errors.Join(err, errInvalidPasswordLen)
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
		err = errors.Join(err, errInvalidPasswordLower)
	}
	if !upper {
		err = errors.Join(err, errInvalidPasswordUpper)
	}
	if !special {
		err = errors.Join(err, errInvalidPasswordSymbol)
	}
	if !number {
		err = errors.Join(err, errInvalidPasswordNumber)
	}

	if err != nil {
		return err
	}

	return err
}

func NewCredentials(username string, password string) *Credentials {
	return &Credentials{username, password}
}
