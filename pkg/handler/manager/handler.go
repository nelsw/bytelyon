package manager

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/manager"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/rs/zerolog/log"
)

const banner = `
██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║
██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║
██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝`

func init() {

	// print banner on startup as sign of life
	fmt.Println(logs.BlueIntense + banner + logs.Default)

	// load the .env file to get app config
	godotenv.Load()

	// init the global logs now that we have an app mode
	logs.Init()

	// init Playwright
	pw.Init()

	// log the app mode and process id for future reference
	log.Info().
		Str("mode", os.Getenv("MODE")).
		Int("pid", os.Getpid()).
		Send()
}

func Start() {

	mgr := manager.New()
	go mgr.Start()

	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("Listening for quit signal (Ctrl+C) ...")
	<-quit
	fmt.Println()
	log.Info().Msg("Quitting ...")

	toCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := mgr.Stop(toCtx); err != nil {
		log.Err(err).Msg("Manager stop failure")
	}

	<-toCtx.Done()
	log.Info().Msg("Exiting ...")
}
