package logs

import (
	"github.com/rs/zerolog/log"
)

func Init() {
	log.Logger = MakeZerolog()
	Meow()
}
