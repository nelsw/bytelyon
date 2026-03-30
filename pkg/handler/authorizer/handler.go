package authorizer

import (
	"errors"
	"strings"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	. "github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var ErrInvalidAuthType = errors.New("invalid authorizer type; must be 'Bearer' or 'Basic'")

func Handler(r Request) (any, error) {

	r.Log()

	tokenType, token, ok := strings.Cut(r.Authorization(), " ")

	if !ok {
		return r.AuthErr(ErrInvalidAuthType), nil
	}

	if tokenType == "Bearer" {
		userID, err := ParseJWT(token)
		if err != nil {
			log.Err(err).Msg("JWT parse failed!")
			return r.AuthErr(err), nil
		}
		log.Debug().Msg("JWT parsed")
		return r.AuthOK(userID, token), nil
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
	} else if userID, err = authenticate(creds.Username, creds.Password); err != nil {
		log.Warn().Err(err).Msg("authentication failed!")
		return r.AuthErr(err), nil
	} else if token, err = NewJWT(userID); err != nil {
		log.Err(err).Msg("JWT creation failed!")
		return r.AuthErr(err), nil
	}

	log.Debug().Msg("authentication successful")

	if r.Query("action") == "login" {
		return r.OK(map[string]any{
			"token":  token,
			"userId": userID.String(),
		}), nil
	}

	return r.AuthOK(userID, token), nil
}

func authenticate(username, password string) (userID ulid.ULID, err error) {

	var email *Email
	if email, err = db.Get(&Email{Address: username}); err != nil {
		log.Warn().Err(err).Msg("email not found")
		return
	}
	log.Debug().Str("email", email.Address).Msg("found email")

	var pass *Password
	if pass, err = db.Get(&Password{UserID: email.UserID}); err != nil {
		log.Warn().Err(err).Msg("password not found")
		return
	}
	log.Debug().Msg("found password")

	if err = pass.Compare(password); err != nil {
		log.Warn().Err(err).Msg("password incorrect")
		return
	}
	log.Debug().Msg("password correct")

	return email.UserID, nil
}
