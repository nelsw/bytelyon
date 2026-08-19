package main

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func makeLogger() zerolog.Logger {
    return zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatLevel: func(a any) string {
			if a == nil || a == "<nil>" {
				a = "   "
			}
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "TRA":
				return "\033[0;36m" + l + "\033[0m"
			case "DEB":
				return "\033[0;35m" + l + "\033[0m"
			case "INF":
				return "\033[0;32m" + l + "\033[0m"
			case "WAR":
				return "\033[0;33m" + l + "\033[0m"
			case "ERR":
				return "\033[0;31m" + l + "\033[0m"
			case "FAT", "PAN":
				return "\033[41m" + "\033[0;37m" + l + "\033[0m"
			default:
				return ""
			}
		},
	}).With().Timestamp().Logger()
}