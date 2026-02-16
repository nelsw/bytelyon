package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/manager"
)

func init() {
	config.Init("Bot Manager")
	db.Init()
}

func main() {

	mgr := manager.New()

	go mgr.Start()

	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println() // Print a newline after the signal is received to escape cmd

	mgr.Stop()
}
