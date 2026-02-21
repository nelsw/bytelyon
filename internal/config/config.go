package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"regexp"
)

var mode, port string

func Mode() string {
	return mode
}

func IsReleaseMode() bool {
	return mode == "release"
}

func IsDebugMode() bool {
	return mode == "debug"
}

func IsTestMode() bool {
	return mode == "test"
}

func Port() int {
	i, _ := strconv.Atoi(port)
	return i
}

func DBLogMode() int {
	switch mode {
	case "release":
		return 1
	case "debug":
		return 2
	case "test":
		return 3
	}
	return 4
}

func Init() {

	if flag.Parsed() {
		return
	}

	flag.StringVar(&mode, "mode", "debug", "The mode of this app")
	flag.StringVar(&port, "port", "8080", "The port to listen on")
	flag.Parse()

	godotenv.Load()
	if str, ok := os.LookupEnv("MODE"); ok {
		mode = str
	}
	if str, ok := os.LookupEnv("PORT"); ok {
		port = str
	}

	if !regexp.MustCompile(`^(debug|release|test)$`).MatchString(mode) {
		panic(fmt.Sprintf("bad mode: [%s] (modes: debug release test)", mode))
	} else if !regexp.MustCompile(`^([1-9][0-9]{1,3})$`).MatchString(port) {
		panic(fmt.Sprintf("bad port: [%s] (ports: 10-9999)", port))
	}

	fmt.Println("\u001B[0;36m" + `
██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║
██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║
██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝
` + "\u001B[0m")
}
