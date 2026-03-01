package db

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	dbClient "github.com/nelsw/bytelyon/internal/client/dynamodb"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/model"
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var db *gorm.DB

var (
	ctx    context.Context
	dbc    *dynamodb.Client
	models = []dbClient.Entity{
		&model.User{},
		&model.Email{},
		&model.Password{},
		&model.SearchBot{},
		&model.SitemapBot{},
		&model.NewsBot{},
	}
)

func Init() {

	ctx = context.Background()
	dbc = Must(dbClient.New())

	migrate()
	seed()
}

func Builder[T any]() gorm.Interface[T] {
	return gorm.G[T](db)
}

func migrate() {

	if config.IsReleaseMode() || !config.MigrateTables() {
		return
	}

	log.Trace().Int("size", len(models)).Msg("migrating tables")

	var wg sync.WaitGroup
	for _, a := range models {
		wg.Go(func() {
			if err := dbClient.DeleteTable(ctx, dbc, a); err != nil {
				return
			}
			if err := dbClient.CreateTable(ctx, dbc, a); err != nil {
				return
			}
		})
	}
	wg.Wait()
}

func seed() {
	if config.IsReleaseMode() || !config.SeedTables() {
		return
	}

	userID := Must(uuid.NewV7())
	dbClient.PutItem(ctx, dbc, &model.User{userID})
	dbClient.PutItem(ctx, dbc, model.NewEmail(userID, "kowalski7012@gmail.com"))
	dbClient.PutItem(ctx, dbc, model.NewPassword(userID, "Demo123!"))
}
