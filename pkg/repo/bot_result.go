package repo

import (
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func BotResults(userID, botID ulid.ULID, botType model.BotType) model.BotResults {

	arr, err := db.Query(&model.BotResult{BotID: botID, Type: botType})
	if err != nil {
		log.Err(err).
			Str("botID", botID.String()).
			Str("botType", botType.String()).
			Msg("bot results query failed")
		return nil
	}

	log.Info().
		Str("botID", botID.String()).
		Str("botType", botType.String()).
		Int("size", len(arr)).
		Msg("bot results found")

	var res model.BotResults
	for _, result := range arr {
		if result.UserID == userID {
			res = append(res, result)
		}
	}

	log.Info().
		Str("botID", botID.String()).
		Str("botType", botType.String()).
		Int("size", len(res)).
		Msg("bot results found for user")

	return res
}
