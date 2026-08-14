package model

import (
	"fmt"
	"strings"
)

type BotType string

func (t *BotType) UnmarshalJSON(payload []byte) error {
	if text := string(payload); text == `"news"` || text == `"search"` || text == `"sitemap"` {
		*t = BotType(strings.ReplaceAll(text, `"`, ""))
		return nil
	}
	return fmt.Errorf("unknown bot type: %s", payload)
}


