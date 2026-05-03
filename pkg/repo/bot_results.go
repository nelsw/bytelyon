package repo

import (
	"errors"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func FindBotResults(userID, botID ulid.ULID, botType model.BotType) []*model.BotResult {

	l := log.With().
		Stringer("user_id", userID).
		Stringer("botId", botID).
		Stringer("botType", botType).
		Logger()

	l.Info().Msg("finding bot results")

	arr, err := db.Query(&model.BotResult{BotID: botID, Type: botType})
	if err != nil {
		l.Err(err).Msg("bot results query failed")
		return nil
	}

	l.Info().
		Int("size", len(arr)).
		Msg("bot results found")

	var res []*model.BotResult
	for _, result := range arr {
		if result.UserID == userID {
			res = append(res, result)
		}
	}

	l.Info().
		Int("size", len(res)).
		Msg("bot results found for user")

	return res
}

func DeleteBotResults(userID, botID ulid.ULID, botType model.BotType) (err error) {

	l := log.With().
		Stringer("userId", userID).
		Stringer("botId", botID).
		Stringer("botType", botType).
		Logger()

	l.Info().Msg("deleting bot results")

	for _, result := range FindBotResults(userID, botID, botType) {
		if result.UserID != userID {
			log.Warn().Msgf("cannot delete bot results for non owner: %s", userID)
			continue
		}
		err = errors.Join(err, db.Delete(result))
	}

	if err != nil {
		l.Err(err).Msg("failed to delete bot results")
	} else {
		l.Info().Msg("bot results deleted")
	}

	return
}
