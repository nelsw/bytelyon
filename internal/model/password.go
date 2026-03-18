package model

import (
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Model
	Hash []byte `json:"-" dynamodbav:"Hash,binary"`
}

// Authenticate returns nil if the given plaint text value is equivalent to this Password.Hash, or an error on failure.
func (p Password) Authenticate(text string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(text))
}

func NewPassword(userID ulid.ULID, text string) *Password {
	return &Password{
		Model: Model{UserID: userID},
		Hash:  Must(bcrypt.GenerateFromPassword([]byte(text), bcrypt.MinCost)),
	}
}

func (p Password) Update(text string) {
	p.Hash = Must(bcrypt.GenerateFromPassword([]byte(text), bcrypt.MinCost))
}
