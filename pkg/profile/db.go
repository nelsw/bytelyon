package profile

import (
	"fmt"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/oklog/ulid/v2"
)

func key(uid ulid.ULID) string { return fmt.Sprintf("users/%s/profile.json", uid) }

func Save(uid ulid.ULID, m *Model) error {
	fmt.Printf("Saving model %+v\n%s\n", m, string(json.Of(m)))
	return s3.Put(key(uid), json.Of(m), false)
}

func Find(uid ulid.ULID) (*Model, error) {
	out, err := s3.Get(key(uid), false)
	if err != nil {
		return nil, err
	}
	var m Model
	if err = json.Unmarshal(out, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
