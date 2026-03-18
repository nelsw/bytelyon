package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"unicode"

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

func Safe[T any](t T, err error) T {
	if err != nil {
		log.Warn().Err(err).Msg("suppressing error because who needs sleep?")
	}
	return t
}

// Name returns the name of the type;
// Helpful for getting a struct name.
func Name(a any) string {
	// get the reflection Type
	t := reflect.TypeOf(a)

	// check if the param is a ptr
	if t.Kind() == reflect.Ptr {
		// if so, return the element type
		t = t.Elem()
	}
	return t.Name()
}

func SplitByCase(s string) []string {
	var words []string
	var currentWord []rune

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			words = append(words, string(currentWord))
			currentWord = []rune{r}
		} else {
			currentWord = append(currentWord, r)
		}
	}
	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}
	return words
}

func TableName(a any) string {
	var n string
	if os.Getenv("MODE") == "release" {
		n = "ByteLyon_"
	} else if os.Getenv("MODE") == "debug" {
		n = "ByteLyon_Debug_"
	} else { // test
		n += "ByteLyon_Test_"
	}
	return n + strings.Join(SplitByCase(Name(a)), "_")
}

func Capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}
