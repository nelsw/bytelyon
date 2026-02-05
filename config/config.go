package config

import (
	"flag"
	"fmt"
	"regexp"
)

// Config defines global configuration properties for this app,
// and populated by either cli args or default values.
type Config struct {
	Mode string
}

// New creates a Config with default values and updates them with given CLI arguments if available.
// It also validates certain Config values and initializes the app logger before returning a Config pointer.
func New(s ...string) *Config {

	var mode string
	if len(s) > 0 {
		mode = s[0]
	}

	if len(flag.Args()) > 0 {
		flag.StringVar(&mode, "mode", "debug", "The Mode of this app")
		flag.Parse()
	}

	if !regexp.MustCompile(`^(debug|release|test)$`).MatchString(mode) {
		panic(fmt.Sprintf("Mode unknown: %s (available Mode: debug release test)", mode))
	}

	return &Config{mode}
}
