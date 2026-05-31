package json

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

func of(a any) (b []byte) {
	if a == nil {
		b = []byte(`{}`)
	} else {
		b, _ = Serialize(a)
	}
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
	a, _ = Deserialize[T](b)
	return
}

func Serialize(a any) ([]byte, error) {
	return json.MarshalIndent(a, "", "  ")
}

func Deserialize[T any](b []byte) (t T, err error) {
	return t, json.Unmarshal(b, &t)
}
