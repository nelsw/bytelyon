package repo

import (
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func DeleteBotResult(userID, botID ulid.ULID, botType model.BotType) (err error) {

	l := log.With().
		Stringer("userID", userID).
		Stringer("botID", botID).
		Stringer("botType", botType).
		Logger()

	l.Info().Msg("deleting bot result")

	for _, result := range FindBotResults(userID, botID, botType) {
		if result.ID == botID {
			err = db.Delete(result)
			break
		}
	}

	if err != nil {
		l.Err(err).Msg("failed to delete bot result")
	} else {
		l.Info().Msg("bot result deleted")
	}

	return
}
