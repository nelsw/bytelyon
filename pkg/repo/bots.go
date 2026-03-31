package repo

import (
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func FindBots(userID ulid.ULID, isReady ...bool) []*model.Bot {

	var arr []*model.Bot
	for _, botType := range model.BotTypes() {
		arr = append(arr, FindBotsByType(userID, botType, isReady...)...)
	}

	log.Info().
		Stringer("userID", userID).
		Int("size", len(arr)).
		Msg("bots found")

	return arr
}

func FindBotsByType(userID ulid.ULID, botType model.BotType, isReady ...bool) model.Bots {

	l := log.With().
		Stringer("user_id", userID).
		Stringer("botType", botType).
		Logger()

	all, err := db.Query(&model.Bot{UserID: userID, Type: botType})
	if err != nil {
		l.Error().Err(err).Msg("bots query failed")
		return nil
	}

	l.Info().
		Int("size", len(all)).
		Msg("find bots")

	// if we are looking for all bots, return do not filter by ready
	if len(isReady) == 0 || !isReady[0] {
		return all
	}

	var ready model.Bots
	for _, b := range all {
		if b.IsReady() {
			ready = append(ready, b)
		}
	}

	l.Info().
		Int("size", len(ready)).
		Msg("bots ready")

	return ready
}
