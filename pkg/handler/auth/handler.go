package auth

import (
	"strings"

	"github.com/nelsw/bytelyon/internal/util"
	. "github.com/nelsw/bytelyon/pkg/api"
	. "github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func Handler(r Request) (AuthResponse, error) {

	r.Log()

	if r.IsPreflight() {
		return r.AuthOK(), nil
	}

	tokenType, token, ok := strings.Cut(r.Authorization(), " ")

	if !ok {
		return r.AuthErr("invalid auth type; must be 'Bearer' or 'Basic'"), nil
	}

	if tokenType == "Bearer" {
		userID, err := ParseJWT(token)
		if err != nil {
			log.Err(err).Msg("JWT parse failed!")
			return r.AuthErr(err), nil
		}
		log.Debug().Msg("JWT parsed")
		return r.AuthOK(userID), nil
	}

	var userID ulid.ULID
	if creds, err := ParseCredentials(token); err != nil {
		log.Debug().Err(err).Msg("credentials invalid")
		return r.AuthErr(err), nil
	} else if err = creds.ValidateUsername(); err != nil {
		log.Debug().Err(err).Msg("username invalid")
		return r.AuthErr(err), nil
	} else if err = creds.ValidatePassword(); err != nil {
		log.Debug().Err(err).Msg("password invalid")
		return r.AuthErr(err), nil
	} else if userID, err = creds.Authenticate(); err != nil {
		log.Warn().Err(err).Msg("authentication failed!")
		return r.AuthErr(err), nil
	} else if token, err = NewJWT(util.Ptr(User{ID: userID})); err != nil {
		log.Err(err).Msg("JWT creation failed!")
		return r.AuthErr(err), nil
	}

	log.Debug().Msg("authentication successful")
	return r.AuthOK(userID), nil
}
