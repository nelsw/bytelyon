package parser

import (
	"bufio"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	contains,
	equals,
	hasPrefix map[string]bool
)

func init() {

	ƒ := func(name string) map[string]bool {

		set := make(map[string]bool)

		file, err := os.Open("skip-contains.txt")
		if err != nil {
			log.Warn().Err(err).Msg("failed to open file")
			return set
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			set[scanner.Text()] = true
		}

		if err = scanner.Err(); err != nil {
			log.Warn().Err(err).Msg("failed to read file")
		}
		return set
	}

	contains = ƒ("skip-contains.txt")
	equals = ƒ("skip-equals.txt")
	hasPrefix = ƒ("skip-has-prefix.txt")
}

func Skip(s string) bool {
	s = strings.ToLower(s)

	for k := range contains {
		if strings.Contains(s, k) {
			return true
		}
	}

	for k := range hasPrefix {
		if strings.HasPrefix(s, k) {
			return true
		}
	}

	_, ok := equals[s]
	return ok
}
