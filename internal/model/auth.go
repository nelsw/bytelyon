package model

import (
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"unicode"
)

var (
	errInvalidEmailAddress   = errors.New("invalid email address")
	errInvalidPasswordLen    = errors.New("password must contain at least 8 characters")
	errInvalidPasswordLower  = errors.New("password must contain at least one lowercase letter")
	errInvalidPasswordNumber = errors.New("password must contain at least one number")
	errInvalidPasswordSymbol = errors.New("password must contain at least one symbol")
	errInvalidPasswordUpper  = errors.New("password must contain at least one uppercase letter")
)

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewBasicAuth(str string) (*Auth, error) {

	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	username, password, ok := strings.Cut(string(b), ":")
	if !ok {
		return nil, errors.New("invalid basic token; must be base64 encoded '<email>:<password>'")
	}

	return &Auth{
		Username: username,
		Password: password,
	}, nil
}

func (a *Auth) Validate() error {

	if _, err := mail.ParseAddress(a.Username); err != nil {
		return errors.Join(err, errInvalidEmailAddress)
	}

	if len(a.Password) < 8 {
		return errInvalidPasswordLen
	}

	var number, lower, upper, special bool
	for _, c := range a.Password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
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

func (a *Auth) Authenticate() {

}
