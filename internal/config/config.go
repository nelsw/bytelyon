package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	. "github.com/nelsw/bytelyon/internal/util"

	"regexp"
)

var banner = "\u001B[0;36m" + `
██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║
██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║
██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝
` + "\u001B[0m"

var cfg = map[string]any{
	"MODE":              os.Getenv("MODE"),
	"PORT":              8085,
	"DB_MIGRATE_TABLES": 0,
	"JWT_SECRET":        "",
}

func Get[T any](key string) T { return cfg[key].(T) }
func IsDebugMode() bool       { return Mode() == "debug" }
func IsReleaseMode() bool     { return Mode() == "release" }
func IsTestMode() bool        { return Mode() == "test" }
func JwtKey() []byte          { return []byte(Get[string]("JWT_SECRET")) }
func MigrateTables() bool     { return Get[int]("DB_MIGRATE_TABLES") == 1 }
func Mode() string            { return Get[string]("MODE") }
func Port() int               { return Get[int]("PORT") }

func Init() {
	if !loadFromCli() {
		loadFromEnv()
	}
	validateCfg()
	fmt.Println(banner)
	fmt.Println("mode: ", Mode())
	fmt.Println()
}

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
		if m, err = godotenv.Read(RootDir(f)); err == nil {
			break
		}
	}

	if err != nil {
		return
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
	if !regexp.MustCompile(`^(debug|release|test)$`).MatchString(Mode()) {
		panic(fmt.Sprintf("bad mode: [%s] (modes: debug release test)", Mode()))
	} else if port := Port(); port < 10 || port > 9999 {
		panic(fmt.Sprintf("bad port: [%d] (ports: 10-9999)", port))
	}
}
