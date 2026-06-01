package json

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

func of(a any) []byte {

	if a == nil {
		return []byte(`{}`)
	}

	b, err := Marshal(a)
	if err != nil {
		log.Warn().Err(err).Any("a", a).Msg("failed to serialize")
		return []byte(`{}`)
	}

	return b
}

func Of(a ...any) (b []byte) {

	if len(a) == 0 {
		return []byte(`{}`)
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

func Marshal(v any) (b []byte, err error) {
	if b, err = json.Marshal(v); err != nil {
		log.Warn().Err(err).Msgf("failed to marshal %T", v)
	}
	return
}

func Unmarshal(b []byte, a any) (err error) {
	if err = json.Unmarshal(b, a); err != nil {
		log.Warn().Err(err).Msgf("failed to unmarshal %T", a)
	}
	return
}
