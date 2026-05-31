package util

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
)

func JSON(a any, pretty ...bool) []byte {

	if a == nil {
		return []byte(`{}`)
	}

	var b []byte

	switch v := a.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		b, _ = json.Marshal(a)
	}

	if len(pretty) == 0 {
		return b
	}

	m := make(map[string]any)
	_ = json.Unmarshal(b, &m)
	b, _ = json.MarshalIndent(m, "", "\t")

	log.Trace().Str("JSON", string(b)).Send()

	return b
}

func Path(a ...any) string {
	var arr []string
	for _, e := range a {
		arr = append(arr, fmt.Sprint(e))
	}
	return path.Join(arr...)
}

func Trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	s = s[:n-3] + "..."
	return s
}

func HasPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
