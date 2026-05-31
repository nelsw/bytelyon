package email

import (
	"encoding/json"
	"fmt"
	"net/mail"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func key(txt string) string { return fmt.Sprintf("emails/%s.json", id.NewUUID(txt)) }

func Create(uid ulid.ULID, txt string) error {
	if err := Validate(txt); err != nil {
		return err
	} else if ok, _ := Exists(txt); ok {
		return fmt.Errorf("email already exists")
	}
	return s3.Put(key(txt), util.JSON(&Model{uid, txt}), false)
}

func Delete(txt string) error {
	if ok, _ := Exists(txt); !ok {
		return nil
	}
	return s3.Delete(key(txt), false)
}

func Exists(txt string) (bool, error) {
	_, err := Find(txt)
	return err == nil, err
}

func Find(txt string) (uid ulid.ULID, err error) {

	var out []byte
	if out, err = s3.Get(key(txt), false); err != nil {
		return
	}

	var m Model
	if err = json.Unmarshal(out, &m); err != nil {
		return
	}
	return m.UID, nil
}

func Validate(txt string) (err error) {
	_, err = mail.ParseAddress(txt)
	return
}
