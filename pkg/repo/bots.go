package repo

import (
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func FindBots(userID ulid.ULID) model.Bots {

	l := log.With().
		Str("ƒ", "FindBots").
		Stringer("userId", userID).
		Logger()

	l.Info().Send()

	var arr []*model.Bot
	for _, botType := range model.BotTypes() {
		arr = append(arr, FindBotsByType(userID, botType)...)
	}

	l.Info().Int("size", len(arr)).Send()

	return arr
}

func FindBotsByType(userID ulid.ULID, botType model.BotType) model.Bots {

	l := log.With().
		Str("ƒ", "FindBotsByType").
		Stringer("userId", userID).
		Stringer("botType", botType).
		Logger()

	l.Info().Send()

	all, err := db.Query(&model.Bot{UserID: userID, Type: botType})
	if err != nil {
		l.Error().Err(err).Msg("bots query failed")
		return nil
	}

	l.Info().Int("size", len(all)).Send()

	return all
}
