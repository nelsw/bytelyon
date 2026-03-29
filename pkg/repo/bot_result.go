package repo

import (
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

func DeleteBotResult(userID, botID ulid.ULID, botType model.BotType) error {
	for _, result := range FindBotResults(userID, botID, botType) {
		if result.ID != botID {
			continue
		}
		return db.Delete(result)
	}
	return nil
}
