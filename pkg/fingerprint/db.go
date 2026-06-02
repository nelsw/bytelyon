package fingerprint

import (
	"errors"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util/json"

	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

func key(uid ulid.ULID, typ bot.Type, tgt string) string {
	return fmt.Sprintf("users/%s/%s/%s/fingerprint.json", uid, typ, tgt)
}

func Find(uid ulid.ULID, typ bot.Type, tgt string) *playwright.OptionalStorageState {

	out, err := s3.Get(key(uid, typ, tgt), false)
	if err != nil {
		return &playwright.OptionalStorageState{}
	}

	var m playwright.OptionalStorageState
	if err = json.Unmarshal(out, &m); err != nil {
		return &playwright.OptionalStorageState{}
	}
	return &m
}

func Save(uid ulid.ULID, typ bot.Type, tgt string, m *playwright.StorageState) error {
	if m == nil {
		return errors.New("nil StorageState")
	} else if err := s3.Put(key(uid, typ, tgt), json.Of(m), false); err != nil {
		return err
	}
	return nil
}
