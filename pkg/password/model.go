package password

import "golang.org/x/crypto/bcrypt"

type Model struct {
	Hash []byte `json:"hash"`
}

// Compare compares a bcrypt hashed password with its possible plaintext equivalent.
func (m *Model) Compare(txt string) error {
	return bcrypt.CompareHashAndPassword(m.Hash, []byte(txt))
}
