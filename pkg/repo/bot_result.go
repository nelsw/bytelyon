package repo

import (
	"errors"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func FindBotResult(userID, botID, ID ulid.ULID, botType model.BotType) (*model.BotResult, error) {

	l := log.With().
		Stringer("userId", userID).
		Stringer("botId", botID).
		Stringer("id", ID).
		Logger()

	l.Info().Msg("Finding bot result")

	res, err := db.Get(&model.BotResult{ID: ID, BotID: botID, Type: botType})
	if err != nil {
		l.Error().Err(err).Msg("failed to find bot result")
		return nil, err
	}
	if res.UserID != userID {
		l.Warn().Msg("found bot result with invalid user ID")
		return nil, errors.New("found bot result with invalid user ID")
	}

	l.Info().Msg("Found bot result")

	return res, nil
}

func DeleteBotResult(userID, botID ulid.ULID, botType model.BotType) error {

	l := log.With().
		Stringer("userId", userID).
		Stringer("botId", botID).
		Stringer("botType", botType).
		Logger()

	l.Info().Msg("deleting bot result")

	if res, err := FindBotResult(userID, botID, botID, botType); err != nil {
		l.Error().Err(err).Msg("failed to find bot result")
		return err
	} else if err = db.Delete(res); err != nil {
		l.Err(err).Msg("failed to delete bot result")
		return err
	}

	l.Info().Msg("bot result deleted")

	return nil
}
