package json

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

func of(a any) (b []byte) {
	if a == nil {
		b = []byte(`{}`)
	} else {
		b, _ = json.MarshalIndent(a, "", "\t")
	}
	log.Trace().Msgf("json\n%v\n%s", a, b)
	return b
}

func Of(a ...any) (b []byte) {

	if len(a) == 0 {
		return of(nil)
	}

	if len(a) == 1 {
		return of(a[0])
	}

	if len(a)%2 != 0 {
		log.Warn().
			Any("args", a).
			Msg("odd number of arguments")
		a = append(a, nil)
	}

	m := make(map[string]any)
	for i := 1; i < len(a); i += 2 {
		m[a[i-1].(string)] = a[i]
	}

	return of(m)
}

func To[T any](b []byte) (a T) {
	_ = json.Unmarshal(b, &a)
	return
}
