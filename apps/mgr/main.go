package main

import (
	"github.com/nelsw/bytelyon/apps/mgr/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
