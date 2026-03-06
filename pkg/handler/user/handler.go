package user

import (
	"github.com/nelsw/bytelyon/internal/service/ses"
	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	. "github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func Handler() {

}

func Login() {

}

func Signup(r Request) Response {

	creds := &Credentials{}

	email, err := db.Get[Email](&Email{})
	if err != nil {
		log.Error().Err(err).Msg("failed to get email on signup")
		return r.BAD(err)
	}

	var user *User
	var userID ulid.ULID
	if email.UserID != ulid.Zero {
		userID = email.UserID
	} else {
		user = NewUser()
		if err = db.Put(user); err != nil {
			log.Error().Err(err).Msg("failed to put user on signup")
			r.BAD(err)
		} else if err = db.Put(&Password{UserID: user.ID}); err != nil {
			log.Error().Err(err).Msg("failed to put password on signup")
			r.BAD(err)
		} else if err = db.Put(&Email{Address: creds.Username, UserID: user.ID}); err != nil {
			log.Error().Err(err).Msg("failed to put email on signup")
			r.BAD(err)
		}
		userID = user.ID
	}

	tkn := NewToken(userID, ConfirmEmailTokenType)
	if err = db.Put(tkn); err != nil {
		log.Error().Err(err).Msg("failed to put token on signup")
		return r.BAD(err)
	}

	if err = ses.SendEmailConfirmation(creds.Username, tkn.UserID.String()); err != nil {
		log.Error().Err(err).Msg("failed to send email on signup")
		return r.BAD(err)
	}

	var jwt string
	if jwt, err = NewJWT(user); err != nil {
		log.Error().Err(err).Msg("failed to generate JWT token")
		return r.BAD(err)
	}

	return r.OK(map[string]any{
		"isAuthenticated": true,
		"context": map[string]any{
			"token": jwt,
		},
	})
}

func ValidateToken() {

}

func ResetPassword(r Request, address string) Response {
	email, err := db.Get[Email](&Email{Address: address})
	if err != nil || email.UserID == ulid.Zero {
		return r.BAD(map[string]string{"error": "failed to find email; try signing up?"})
	}

	tkn := NewToken(email.UserID, ResetPasswordTokenType)
	if err = db.Put(tkn); err != nil {
		return r.BAD(map[string]string{"error": "failed to save forgot password token"})
	} else if err = ses.SendPasswordReset(address, tkn.UserID.String()); err != nil {
		return r.BAD(map[string]string{"error": "failed to send email on forgot password"})
	}

	return r.OK(map[string]string{"message": "password reset email sent"})
}
