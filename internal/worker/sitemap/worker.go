package sitemap

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	client "github.com/nelsw/bytelyon/internal/client/dynamodb"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	context.Context
	*dynamodb.Client
	*model.SitemapBot
}

func New(ctx context.Context, dbc *dynamodb.Client, bot *model.SitemapBot) *Worker {
	return &Worker{ctx, dbc, bot}
}

func (w *Worker) Work() {

	m := NewMapper(&fetcher{}, w.Target)
	m.Add()
	m.Map(w.Target, 3)
	m.Wait()

	sort.Strings(m.Relative())
	sort.Strings(m.Remote())

	err := client.PutItem(w.Context, w.Client, &model.SitemapBotData{
		BotID:    w.BotID,
		URL:      w.Target,
		Domain:   util.Domain(w.Target),
		Relative: m.Relative(),
		Remote:   m.Remote(),
	})

	if err != nil {
		log.Err(err).Msg("Failed to create sitemap")
	}
}
