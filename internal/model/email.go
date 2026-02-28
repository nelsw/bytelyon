package model

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Email represents a single mail address, the User it belongs to, a token for address confirmation.
type Email struct {

	// ID is a unique email address and primary key of the Email table.
	ID string `json:"address" dynamodbav:"ID,binary"`

	// UserID is a foreign key reference to define which Email belongs-to a User.
	UserID uuid.UUID `json:"-" dynamodbav:"UserID,binary"`

	// Token is present when confirming an Email address, and omitted (nil) once empty.
	Token string `json:"token,omitempty" dynamodbav:"Token,omitempty"`
}

// Validate asserts that the given address follows RFC 5322.
func (e *Email) Validate() error {
	log.Trace().Str("email", e.ID).Msg("validating email address")
	if _, err := mail.ParseAddress(e.ID); err != nil {
		log.Warn().Err(err).Str("email", e.ID).Msg("invalid email address")
		return err
	}
	log.Debug().Str("email", e.ID).Msg("valid email address")
	return nil
}
