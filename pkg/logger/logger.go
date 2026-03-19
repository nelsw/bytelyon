package logger

import (
	"github.com/rs/zerolog/log"
)

func Init() {
	log.Logger = MakeZerolog()
}
