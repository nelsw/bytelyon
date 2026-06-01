package password

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"golang.org/x/crypto/bcrypt"
)

func key(pid uuid.UUID) string { return fmt.Sprintf("porkys/%s.json", pid) }

func Create(txt string) (u uuid.UUID, err error) {
	if err = Validate(txt); err != nil {
		return
	}

	var b []byte
	if b, err = bcrypt.GenerateFromPassword([]byte(txt), bcrypt.MinCost); err != nil {
		return
	}

	u = id.NewUUID()
	return u, s3.Put(key(u), json.Of(&Model{b}), false)
}

func Find(pid uuid.UUID) (*Model, error) {
	out, err := s3.Get(key(pid), false)
	if err != nil {
		return nil, err
	}

	var m Model
	if err = json.Unmarshal(out, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func Validate(password string) error {

	if len(password) < 8 {
		return errors.New("password must contain at least 8 characters")
	}

	var number, lower, upper, special bool
	for _, r := range password {
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
		return errors.New("password must contain at least one lowercase letter")
	} else if !upper {
		return errors.New("password must contain at least one uppercase letter")
	} else if !special {
		return errors.New("password must contain at least one symbol")
	} else if !number {
		return errors.New("password must contain at least one number")
	}

	return nil
}
