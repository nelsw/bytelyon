package bot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nelsw/bytelyon/pkg/news"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/search"
	"github.com/nelsw/bytelyon/pkg/sitemap"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/nelsw/bytelyon/pkg/util/urls"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func key(uid ulid.ULID, typ Type, tgt string) string {
	return fmt.Sprintf("users/%s/%s/%s/config.json", uid, typ, tgt)
}

func Delete(uid ulid.ULID, typ Type, tgt string) (err error) {
	switch typ {
	case News:
		err = news.Delete(uid, tgt)
	case Search:
		err = search.Delete(uid, tgt)
	case Sitemap:
		err = sitemap.Delete(uid, tgt)
	}
	if err != nil {
		return
	}
	return s3.Delete(key(uid, typ, tgt), false)
}

func Find(uid ulid.ULID, typ Type) Models {

	var mm Models
	prefix := fmt.Sprintf("users/%s/%s/", uid, typ)
	arr, _ := s3.ListDirectories(prefix)
	for _, k := range arr {

		if !strings.HasSuffix(k, "/config.json") {
			continue
		}

		b, err := s3.Get(k, false)
		if err != nil {
			log.Warn().Err(err).Str("key", k).Msg("failed to get config")
			continue
		}

		mm = append(mm, json.To[*Model](b))
	}

	sort.Sort(mm)

	return mm
}

func Save(uid ulid.ULID, m *Model) error {

	if err := m.Validate(); err != nil {
		return err
	}

	if m.Type == Sitemap {
		m.Target = urls.Domain(m.Target)
	} else {
		m.Target = strings.ToLower(m.Target)
	}

	return s3.Put(key(uid, m.Type, m.Target), json.Of(m), false)
}
