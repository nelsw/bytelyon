package user

import (
	"github.com/nelsw/bytelyon/pkg/email"
	"github.com/nelsw/bytelyon/pkg/password"
	"github.com/oklog/ulid/v2"
)

func Login(user, pass string) (um *Model, err error) {
	var uid ulid.ULID
	var pm *password.Model
	if err = email.Validate(user); err != nil {
		return
	} else if err = password.Validate(pass); err != nil {
		return
	} else if uid, err = email.Find(user); err != nil {
		return
	} else if um, err = Find(uid); err != nil {
		return
	} else if pm, err = password.Find(um.PID); err != nil {
		return
	} else if err = pm.Compare(pass); err != nil {
		return
	}
	return
}
