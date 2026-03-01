package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/internal/util"

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
	"MODE": "debug",
	"PORT": 8085,
}

func Aws() aws.Config         { return cfg["AWS_CONFIG"].(aws.Config) }
func Get[T any](key string) T { return cfg[key].(T) }
func JwtKey() []byte          { return []byte(cfg["JWT_SECRET"].(string)) }
func Mode() string            { return Get[string]("MODE") }
func ModeTitle() string       { return strings.ToUpper(Mode()[0:1]) + Mode()[1:] }
func IsReleaseMode() bool     { return Mode() == "release" }
func IsDebugMode() bool       { return Mode() == "debug" }
func IsTestMode() bool        { return Mode() == "test" }
func Port() int               { return Get[int]("PORT") }
func MigrateTables() bool     { return Get[int]("DB_MIGRATE_TABLES") == 1 }
func SeedTables() bool        { return Get[int]("DB_SEED_TABLES") == 1 }

func Init() {

	if !loadFromCli() {
		loadFromEnv()
	}

	loadAwsConfig()
	validateCfg()

	fmt.Println(banner)
}

func loadAwsConfig() {
	c, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	cfg["AWS_CONFIG"] = c
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
		if m, err = godotenv.Read(util.RootDir(f)); err == nil {
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
