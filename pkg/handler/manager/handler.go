package manager

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/job"
	logger2 "github.com/nelsw/bytelyon/pkg/logger"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

const banner = `
██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║
██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║
██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝`

var (
	ctx context.Context
	DB  *dynamodb.Client
	S3  *s3.Client
)

func init() {

	// print banner on startup as sign of life
	fmt.Println(logger2.BlueIntense + banner + logger2.Default)

	// load the .env file to get app config
	godotenv.Load()

	// init the global logger now that we have an app mode
	log.Logger = logger2.MakeZerolog()

	// log the app mode and process id for future reference
	log.Info().
		Str("mode", os.Getenv("MODE")).
		Int("pid", os.Getpid()).
		Send()

	// install playwright
	if err := playwright.Install(&playwright.RunOptions{Logger: logger2.NewSlog()}); err != nil {
		log.Fatal().Err(err).Msg("failed to install playwright")
	}
	// define a root context
	ctx = context.Background()

	// init aws config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load aws config")
	}

	// define dynamodb and s3 clients
	DB = dynamodb.NewFromConfig(cfg)
	S3 = s3.NewFromConfig(cfg)
}

func Start() {

	mgr := &Manager{}
	go mgr.Start()

	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("Listening for quit signal (Ctrl+C) ...")
	<-quit

	log.Info().Msg("Quitting ...")

	toCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := mgr.Stop(toCtx); err != nil {
		log.Err(err).Msg("Manager stop failure")
	}

	<-toCtx.Done()
	log.Info().Msg("Exiting ...")
}

type Manager struct {
	stop, done bool
}

func (m *Manager) Start() {

	log.Info().Msg("bot manager looking for work")

	for !m.stop {

		log.Info().Msg("bot manager working")

		m.done = false
		m.work()
		m.done = true

		log.Info().Msg("bot manager work complete")

		if m.stop {
			return
		}

		d := time.Duration(15) * time.Second

		log.Debug().
			Dur("duration", d).
			Msg("bot manager sleeping")

		time.Sleep(d)
	}
}

func (m *Manager) Stop(ctx context.Context) error {

	m.stop = true

	timer := time.NewTimer(time.Second)

	defer func() {
		timer.Stop()
		log.Debug().Msg("bot manager stopped")
	}()

	log.Info().Msg("bot manager stopping")
	for {
		if m.done {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(time.Second)
		}
	}
}

func (m *Manager) work() {

	log.Info().Msg("bot manager looking for users")

	users, err := db.Scan(&model.User{})
	if err != nil {
		log.Error().Err(err).Msg("user scan failed")
		return
	}

	log.Info().Int("size", len(users)).Msg("users found")

	var ready []*model.Bot
	for _, user := range users {

		var found []*model.Bot

		for _, botType := range model.BotTypes() {
			found, err = db.Query(&model.Bot{UserID: user.ID, Type: botType})
			if err != nil {
				log.Error().Err(err).Msg("bot query failed")
				continue
			}
			found = append(found, found...)
		}

		log.Info().
			Int("size", len(found)).
			Stringer("userID", user.ID).
			Msg("bots found")

		for _, bot := range found {
			if bot.IsReady() {
				ready = append(ready, bot)
			}
		}

		log.Info().
			Int("size", len(ready)).
			Stringer("userID", user.ID).
			Msg("bots ready")
	}

	log.Info().
		Int("size", len(ready)).
		Msg("users bots ready for work")

	if len(ready) == 0 {
		log.Info().Msg("no bots ready for work, sleeping ...")
		return
	}

	log.Info().Int("size", len(ready)).Msg("bots ready for work")
	var wg sync.WaitGroup
	for _, bot := range ready {
		wg.Go(func() { job.New(ctx, DB, S3, bot).Work() })
	}
	log.Info().Msg("bots working")
	wg.Wait()
	log.Info().Msg("bots worked")
}
