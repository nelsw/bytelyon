package util

import (
	"encoding/json"
	"fmt"
	"path"
)

func JSON(a any) []byte {
	if a == nil {
		return []byte(`{}`)
	}
	switch v := a.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return Safe(json.Marshal(a))
	}
}

func PrettyJSON(a any) string {
	return string(Safe(json.MarshalIndent(a, "", "\t")))
}

func PrintlnPrettyJSON(a any) {
	fmt.Println(PrettyJSON(a))
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
