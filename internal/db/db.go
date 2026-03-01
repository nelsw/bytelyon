package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	dbClient "github.com/nelsw/bytelyon/internal/client/dynamodb"
	s3Client "github.com/nelsw/bytelyon/internal/client/s3"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/model"
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

var (
	ctx    context.Context
	dbc    *dynamodb.Client
	s3c    *s3.Client
	models = []dbClient.Entity{
		&model.User{},
		&model.Email{},
		&model.Password{},
	}
)

func Init(args ...context.Context) {

	db = Must(gorm.Open(sqlite.Open(BinDir(config.Mode()+".sqlite")), &gorm.Config{
		Logger: logger.NewGorm(),
	}))

	if len(args) > 0 {
		ctx = args[0]
	} else {
		ctx = context.Background()
	}

	dbc = Must(dbClient.New())
	s3c = Must(s3Client.New())

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

	for _, a := range models {
		if err := dbClient.DeleteTable(ctx, dbc, a); err != nil {
			continue
		}
		if err := dbClient.CreateTable(ctx, dbc, a); err != nil {
			continue
		}
	}
}

func seed() {
	if config.IsReleaseMode() || !config.SeedTables() {
		return
	}

	user := &model.User{Must(uuid.NewV7())}
	dbClient.PutItem(ctx, dbc, user)
	dbClient.PutItem(ctx, dbc, model.NewEmail(user, "kowalski7012@gmail.com"))
	dbClient.PutItem(ctx, dbc, model.NewPassword(user, "Demo123!"))
}
