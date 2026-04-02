package main

import (
	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/pw"
)

func init() {
	godotenv.Load()
	logs.Init()
	pw.Init()
}

func main() {

	headless, headed, err := pw.NewBrowsers()
	if err != nil {
		panic(err)
	}

	defer func() {
		headless.Close()
		headed.Close()
		pw.Client.Stop()
	}()

}
