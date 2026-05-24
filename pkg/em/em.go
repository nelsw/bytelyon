package em

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/s3"
)

type Entity interface {
	Key() string
	Val() []byte
	Associations() []Entity
}

func Save(e Entity) error {
	if out, err := json.Marshal(e); err != nil {
		return err
	} else {
		return s3.PutPrivateObject(e.Key(), out)
	}
}

func Find(e Entity) error {
	if out, err := s3.GetPrivateObject(e.Key()); err != nil {
		return err
	} else {
		return json.Unmarshal(out, e)
	}
}

func Delete(e Entity) error {
	return s3.DeletePrivateObject(e.Key())
}

func FindOrCreate(e Entity) error {
	if err := Find(e); err != nil {
		return Save(e)
	}
	return nil
}
