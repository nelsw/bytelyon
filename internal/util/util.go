package util

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
)

const app = "ByteLyon"

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

func RootDir(parts ...string) string {
	dir := Must(os.Getwd())
	for !strings.HasSuffix(dir, "bytelyon") {
		dir = dir[:strings.LastIndex(dir, "/")]
	}
	if len(parts) > 0 {
		dir += "/" + strings.Join(parts, "/")
	}
	return dir
}

func BinDir(parts ...string) string {
	dir := RootDir("bin")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		Check(os.Mkdir(dir, fs.ModePerm))
	}
	if len(parts) > 0 {
		dir += "/" + strings.Join(parts, "/")
	}
	return dir
}

func Name(a any) string {
	t := reflect.TypeOf(a)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func Capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func AppName(args ...string) string {

	if len(args) == 0 {
		return app
	}

	if len(args)%2 == 0 {
		return app + strings.Join(args, "")
	}

	return app + args[0] + strings.Join(args[1:], args[0])
}
