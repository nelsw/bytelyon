package model

import (
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
)

var (
	invalidCredentialsErr    = errors.New("invalid basic token; must be base64 encoded '<email>:<password>'")
	invalidEmailAddressErr   = errors.New("invalid email address")
	invalidPasswordLenErr    = errors.New("password must contain at least 8 characters")
	invalidPasswordLowerErr  = errors.New("password must contain at least one lowercase letter")
	invalidPasswordNumberErr = errors.New("password must contain at least one number")
	invalidPasswordSymbolErr = errors.New("password must contain at least one symbol")
	invalidPasswordUpperErr  = errors.New("password must contain at least one uppercase letter")
)

type Credentials struct {
	Username string
	Password string
}

func ParseCredentials(s string) (*Credentials, error) {

	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Warn().Err(err).Msg("credential encoding invalid")
		return nil, err
	}
	log.Trace().Msg("credential encoding valid")

	username, password, ok := strings.Cut(string(b), ":")
	if !ok {
		log.Warn().Msg("credential format invalid")
		return nil, invalidCredentialsErr
	}
	log.Trace().Msg("credential format valid")

	return &Credentials{username, password}, nil
}

func (c *Credentials) Validate() error {
	if err := c.ValidateUsername(); err != nil {
		log.Warn().Err(err).Msg("username invalid")
	} else if err = c.ValidatePassword(); err != nil {
		log.Warn().Err(err).Msg("password invalid")
	}
	return nil
}

func (c *Credentials) ValidateUsername() error {

	log.Trace().
		Str("email", c.Username).
		Msg("validating email address")

	if _, err := mail.ParseAddress(c.Username); err != nil {
		log.Warn().Err(err).Str("email", c.Username).Msg("invalid email address")
		return errors.Join(err, invalidEmailAddressErr)
	}

	log.Debug().Str("email", c.Username).Msg("valid email address")

	return nil
}

func (c *Credentials) ValidatePassword() error {

	log.Trace().Msg("validating password")

	if len(c.Password) < 8 {
		return invalidPasswordLenErr
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
		return invalidPasswordLowerErr
	} else if !upper {
		return invalidPasswordUpperErr
	} else if !special {
		return invalidPasswordSymbolErr
	} else if !number {
		return invalidPasswordNumberErr
	}
	log.Trace().Msg("password valid")
	return nil
}
