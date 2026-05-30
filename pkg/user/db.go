package user

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func key(uid ulid.ULID) string { return fmt.Sprintf("users/%s/_.json", uid) }

func Find(a any) (m *Model, err error) {

	if a == nil {
		return
	}

	var uid ulid.ULID
	if s, ok := a.(string); ok {
		uid = id.ParseULID(s)
	} else {
		uid, _ = a.(ulid.ULID)
	}

	var out []byte
	if out, err = s3.Get(key(uid), false); err != nil {
		return
	} else if err = json.Unmarshal(out, &m); err != nil {
		return
	}

	m.UID = uid

	return
}

func Save(m *Model) error {
	return s3.Put(key(m.UID), util.JSON(m), false)
}
