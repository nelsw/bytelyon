package config

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/logger"
)

// Config defines global configuration properties for this app,
// and populated by either cli args or default values.
type Config struct {
	Mode string
	Port int
}

// New creates a Config with default values and updates them with given CLI arguments if available.
// It also validates certain Config values and initializes the app logger before returning a Config pointer.
func New(s ...string) *Config {

	c := &Config{gin.DebugMode, 8080}

	if len(s) > 0 {
		c.Mode = s[0]
	}

	if len(flag.Args()) > 0 {
		flag.StringVar(&c.Mode, "mode", "debug", "The Mode of this app")
		flag.IntVar(&c.Port, "port", 8080, "The port to listen on")
		flag.Parse()
	}

	if !regexp.MustCompile(`^(debug|release|test)$`).MatchString(c.Mode) {
		panic(fmt.Sprintf("Mode unknown: %s (available Mode: debug release test)", c.Mode))
	}

	logger.Init(c.Mode)

	return c
}
