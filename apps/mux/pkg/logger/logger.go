package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	Cyan    = "\033[0;36m"
	Purple  = "\033[0;35m"
	Green   = "\033[0;32m"
	Yellow  = "\033[0;33m"
	Red     = "\033[0;31m"
	RedBG   = "\033[41m"
	White   = "\033[0;37m"
	Default = "\033[0m"
)

func init() {
	log.Logger = *New()
}

type builder struct {
	level  zerolog.Level
	fields []string
	caller bool
}

func New(a ...any) *zerolog.Logger {
	b := builder{
		level: zerolog.DebugLevel,
	}
	for _, v := range a {
		switch v := v.(type) {
		case zerolog.Level:
			b.level = v
		case []string:
			b.fields = v
		case bool:
			b.caller = v
		}
	}
	return b.build()
}

func (b builder) build() *zerolog.Logger {

	l := zerolog.New(zerolog.ConsoleWriter{
		Out:         os.Stdout,
		FieldsOrder: b.fields,
		FormatLevel: func(a any) string {
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "TRA":
				return Cyan + l + Default
			case "DEB":
				return Purple + l + Default
			case "INF":
				return Green + l + Default
			case "WAR":
				return Yellow + l + Default
			case "ERR":
				return Red + l + Default
			case "FAT", "PAN":
				return RedBG + White + l + Default
			default:
				return Default + l + Default
			}
		},
	}).Level(b.level)

	if b.caller {
		l = l.With().Caller().Logger()
	}

	return &l
}
