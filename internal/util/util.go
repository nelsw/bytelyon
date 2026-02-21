package util

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func Ptr[T any](a T) *T { return &a }

func Between[T int | float64](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}

func Domain(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	s = strings.Split(s, "/")[0]
	for strings.Count(s, ".") > 1 {
		s = strings.Split(s, ".")[1]
	}
	return s
}

func PrettyPrintln(a any) {
	b, _ := json.MarshalIndent(a, "", "  ")
	fmt.Println(string(b))
}

func Check(err error) {
	if err != nil {
		log.Panic().Err(err).Stack().Send()
	}
}

func Must[T any](t T, err error) T {
	Check(err)
	return t
}

func RootDir() string {
	dir := Must(os.Getwd())
	for !strings.HasSuffix(dir, "bytelyon") {
		dir = dir[:strings.LastIndex(dir, "/")]
	}

	return dir
}

func BinDir(parts ...string) string {
	dir := RootDir() + "/bin"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		Check(os.Mkdir(dir, fs.ModePerm))
	}

	return dir + strings.Join(parts, "/")
}
