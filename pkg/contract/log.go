package contract

import "github.com/rs/zerolog"

type Loggable interface {
	MarshalZerologObject(evt *zerolog.Event)
}
