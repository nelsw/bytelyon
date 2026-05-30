package password

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"golang.org/x/crypto/bcrypt"
)

func key(pid uuid.UUID) string { return fmt.Sprintf("porkys/%s.json", pid) }

// Compare compares a bcrypt hashed password with its possible plaintext equivalent.
func (m *Model) Compare(txt string) error {
	return bcrypt.CompareHashAndPassword(m.Hash, []byte(txt))
}

func Create(txt string) (u uuid.UUID, err error) {
	if err = Validate(txt); err != nil {
		return
	}

	var b []byte
	if b, err = bcrypt.GenerateFromPassword([]byte(txt), bcrypt.MinCost); err != nil {
		return
	}

	u = id.NewUUID()
	return u, s3.Put(key(u), util.JSON(&Model{b}), false)
}

func Delete(pid uuid.UUID, txt string) error {
	if m, err := Find(pid); err != nil {
		return err
	} else if err = m.Compare(txt); err != nil {
		return err
	}
	return s3.Delete(key(pid), false)
}

func Find(pid uuid.UUID) (m *Model, err error) {
	var out []byte
	if out, err = s3.Get(key(pid), false); err == nil {
		err = json.Unmarshal(out, &m)
	}
	return
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
