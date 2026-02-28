package model

import (
	"errors"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	UserID uuid.UUID `json:"-" dynamodbav:"UserID,binary"`
	Hash   []byte    `json:"-" dynamodbav:"Hash,binary"`
	Text   string    `json:"-" dynamodbav:"-"`
}

func (p *Password) Validate() (err error) {

	if len(p.Text) < 8 {
		err = errors.Join(err, errInvalidPasswordLen)
	}

	var number, lower, upper, special bool
	for _, c := range p.Text {
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

	p.Hash, err = bcrypt.GenerateFromPassword([]byte(p.Text), bcrypt.DefaultCost)

	return
}

// Authenticate returns nil if the given plaint text value is equivalent to this Password.Hash, or an error on failure.
func (p *Password) Authenticate(text string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(text))
}
