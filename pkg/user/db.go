package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util/json"

	"github.com/oklog/ulid/v2"
)

func key(uid ulid.ULID) string { return fmt.Sprintf("users/%s/user.json", uid) }

func IDs() (ids []ulid.ULID) {
	prefix := "users/"
	arr, _ := s3.ListDirectories(prefix)
	for _, k := range arr {
		if !strings.HasSuffix(k, "/user.json") {
			continue
		}
		s := strings.TrimPrefix(k, prefix)
		s = strings.TrimSuffix(s, "/user.json")
		ids = append(ids, id.ParseULID(s))
	}

	return
}

func Find(a any) (*Model, error) {

	if a == nil {
		return nil, errors.New("no user id provided")
	}

	var uid ulid.ULID
	if s, ok := a.(string); ok {
		uid = id.ParseULID(s)
	} else {
		uid, _ = a.(ulid.ULID)
	}

	out, err := s3.Get(key(uid), false)
	if err != nil {
		return nil, err
	}

	var m Model
	if err = json.Unmarshal(out, &m); err != nil {
		return nil, err
	}

	m.ID = uid

	return &m, nil
}

func Save(m *Model) error {
	return s3.Put(key(m.ID), json.Of(m), false)
}
