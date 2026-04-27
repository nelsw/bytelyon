package worker

import (
	"flag"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/manager"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func init() {

	// load the .env file to get app config
	godotenv.Load()

	// init the global logs now that we have an app mode
	logs.Init()

	// init Playwright
	pw.Init()

	// print ByteLyon banner to show we're ready
	logs.PrintWorkerBanner()

	// log the app mode and process id for future reference
	log.Info().
		Str("mode", os.Getenv("MODE")).
		Int("pid", os.Getpid()).
		Send()
}

func Run() {

	var u, t string
	var h, f, x bool
	flag.StringVar(&u, "user", "01KMXGBJJE2GMCA1A9EXDGF4AJ", "user id")
	flag.StringVar(&t, "type", model.NewsBotType.String(), "bot type [news, search, sitemap]")
	flag.BoolVar(&h, "headless", true, "use headless browser (will not commandeer your browsers)")
	flag.BoolVar(&f, "force", false, "work a bot even if it's not ready")
	flag.BoolVar(&x, "async", false, "work is executed concurrently on multiple threads")
	flag.Parse()

	log.Info().Bool("headless", h).Msg("instantiating browser ...")
	bro, err := pw.NewBrowser(h)
	if err != nil {
		log.Panic().Err(err).Msg("failed to instantiate browser")
	}
	defer func() {
		bro.Close()
		pw.Client.Stop()
	}()

	var userID ulid.ULID
	if userID, err = ulid.ParseStrict(u); err != nil {
		log.Panic().Err(err).Msg("userID is invalid")
	}
	log.Info().Msg("userID looks OK")

	botType := model.BotType(t)
	if err = botType.Validate(); err != nil {
		log.Panic().Err(err).Msg("BotType is invalid")
	}
	log.Info().Msg("BotType looks OK")

	log.Info().Bool("force work", f).Bool("run async", x).Send()

	bots := repo.FindBotsByType(userID, botType)
	if len(bots) == 0 {
		log.Info().Msg("no bots found")
		return
	}

	var jobs []*manager.Job
	for _, bot := range bots {

		if !bot.IsReady() {
			log.Info().Str("target", bot.Target).Msg("bot is not ready to work")
			if !f {
				log.Debug().Msg("skipping bot")
				continue
			}
			log.Info().Msg("forcing work")
		}

		var ctx playwright.BrowserContext
		if bot.Fingerprint == nil {
			bot.Fingerprint = model.NewFingerprint()
		}
		state := bot.Fingerprint.GetState()

		ctx, err = client.NewContext(bro, state)

		jobs = append(jobs, manager.NewJob(bot, ctx))
	}
	log.Info().Msgf("jobs found: %d", len(jobs))

	if !x {
		log.Info().Msg("running jobs sequentially ...")
		for _, j := range jobs {
			j.Work()
		}
		log.Info().Msg("worked jobs")
		return
	}

	log.Info().Msg("running jobs concurrently ...")
	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Go(j.Work)
	}
	log.Info().Msg("waiting for jobs to finish ...")
	wg.Wait()
	log.Info().Msg("worked jobs")
}
