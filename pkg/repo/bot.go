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

	l.Info().Msg("find bot")

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

	l := log.With().
		Stringer("userID", userID).
		Str("target", target).
		Stringer("botType", botType).
		Logger()

	l.Info().Msg("deleting bot")

	bot, err := FindBot(userID, target, botType)
	if err != nil {
		l.Err(err).Msg("failed to find bot to delete")
		return err
	}

	if err = db.Delete(bot); err != nil {
		l.Err(err).Msg("failed to delete bot")
		return err
	}

	return DeleteBotResults(userID, bot.ID, botType)
}
