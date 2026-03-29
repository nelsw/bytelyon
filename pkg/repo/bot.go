package repo

import (
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func FindBot(userID ulid.ULID, target string, botType model.BotType) (*model.Bot, error) {

	l := log.With().
		Stringer("userID", userID).
		Str("target", target).
		Stringer("botType", botType).
		Logger()

	l.Debug().Msg("find bot")

	bot, err := db.Get(&model.Bot{
		UserID: userID,
		Target: target,
		Type:   botType,
	})

	if err != nil {
		l.Warn().Err(err).Msg("failed find bot")
		return nil, err
	}

	l.Info().Msg("found bot")

	return bot, nil
}

func DeleteBot(userID ulid.ULID, target string, botType model.BotType) error {

	bot, err := FindBot(userID, target, botType)
	if err != nil {
		return err
	}

	if err = DeleteBotResults(userID, bot.ID, botType); err != nil {
		return err
	}

	return db.Delete(bot)
}
