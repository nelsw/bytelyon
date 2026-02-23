package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/internal/util"

	"regexp"
)

var cfg map[string]any

func Get[T any](key string) T { return cfg[key].(T) }
func Mode() string            { return Get[string]("MODE") }
func IsReleaseMode() bool     { return Mode() == "release" }
func IsDebugMode() bool       { return Mode() == "debug" }
func IsTestMode() bool        { return Mode() == "test" }
func Port() int               { return Get[int]("PORT") }

func loadFromCli() bool {

	if len(flag.Args()) == 0 {
		return false
	}

	cfg["MODE"] = *flag.String("mode", "debug", "The mode of this app")
	cfg["PORT"] = *flag.Int("port", 8085, "The port to listen on")
	flag.Parse()

	return true
}

func loadFromEnv() {

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "test"
	}

	files := []string{
		".env." + env + ".local",
		".env.local",
		".env." + env,
		".env",
	}

	var m map[string]string
	var err error
	for _, f := range files {
		if m, err = godotenv.Read(util.RootDir(f)); err == nil {
			break
		}
	}

	var i int
	for k, v := range m {
		if i, err = strconv.Atoi(v); err == nil {
			cfg[k] = i
			continue
		}
		cfg[k] = v
	}
}

func validateCfg() {

	keys := []string{
		"MODE",
		"PORT",
	}

	for _, k := range keys {
		if _, ok := cfg[k]; !ok {
			panic(fmt.Sprintf("missing config key: [%s]", k))
		}
	}

	if !regexp.MustCompile(`^(debug|release|test)$`).MatchString(Mode()) {
		panic(fmt.Sprintf("bad mode: [%s] (modes: debug release test)", Mode()))
	} else if port := Port(); port < 10 || port > 9999 {
		panic(fmt.Sprintf("bad port: [%d] (ports: 10-9999)", port))
	}
}

func Init() {

	if len(cfg) > 0 {
		return
	}

	cfg = make(map[string]any)

	if !loadFromCli() {
		loadFromEnv()
	}

	validateCfg()

	fmt.Println("\u001B[0;36m" + `
██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║
██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║
██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝
` + "\u001B[0m")
}
