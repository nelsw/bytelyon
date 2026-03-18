package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/nelsw/bytelyon/internal/service/ses"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func ProcessToken(c *gin.Context) {

	ID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		badRequest(c, err)
		return
	}

	var tkn Token
	if tkn, err = db.Find[Token](Data{"UserID": ID}); err != nil {
		errRequest(c, err)
		return
	}
	if tkn.UserID == ulid.Zero {
		badRequest(c, "token not found")
		return
	}
	if tkn.Expiry.Before(time.Now()) {
		badRequest(c, "token expired")
		if err = db.Wipe(Token{}, Data{"UserID": ID}); err != nil {
			log.Warn().Err(err).Msg("failed to delete token") // just warn the logs
		}
		return
	}

	if tkn.Type == ConfirmEmailTokenType {

		var jwt string
		if jwt, err = NewJWT(tkn.UserID); err != nil {
			errRequest(c, err)
			return
		}

		if err = db.Wipe(Token{}, Data{"UserID": ID}); err != nil {
			errRequest(c, errors.Join(err, fmt.Errorf("failed to delete token")))
			return
		}

		c.JSON(http.StatusOK, Data{"isAuthenticated": true, "context": Data{"token": jwt}})
		return
	}

	if tkn.Type == ResetPasswordTokenType {

		var pass Password
		if pass, err = db.Find[Password](Data{"UserID": tkn.UserID}); err != nil {
			badRequest(c, "failed to find user and password from token")
			return
		}

		pass.Update(creds(c).Password)
		if err = db.Save(&pass); err != nil {
			errRequest(c, err)
			return
		}

		var jwt string
		if jwt, err = NewJWT(tkn.UserID); err != nil {
			errRequest(c, err)
			return
		}

		if err = db.Wipe(Token{}, Data{"UserID": ID}); err != nil {
			log.Warn().Err(err).Msg("failed to delete token") // just warn the logs
		}

		c.JSON(http.StatusOK, Data{"isAuthenticated": true, "context": Data{"token": jwt}})
		return
	}

	badRequest(c, "invalid token type")
}

func LoginUser(c *gin.Context) { tokenResponse(c, c.MustGet("JWT")) }

func SignupUser(c *gin.Context) {

	// have we seen this email address before
	email, err := db.Find[Email](Data{"Address": creds(c).Username})
	if err != nil {
		errRequest(c, "failed to get email on signup")
		return
	}

	var model Model
	if email.UserID != ulid.Zero {
		model = email.Model
	} else {
		model = createNewUser(c, creds(c))
	}

	tkn := NewConfirmEmailToken(model.UserID)
	if err = db.Save(tkn); err != nil {
		errRequest(c, "failed to put token on signup")
		return
	}

	if err = ses.SendEmailConfirmation(creds(c).Username, tkn.UserID.String()); err != nil {
		errRequest(c, "failed to send email on signup")
		return
	}

	var jwt string
	if jwt, err = NewJWT(model.UserID); err != nil {
		errRequest(c, "failed to generate JWT token")
	} else {
		tokenResponse(c, jwt)
	}
}

func createNewUser(c *gin.Context, creds *Credentials) Model {
	model := Make()
	if err := db.Save(NewUser(model.UserID)); err != nil {
		errRequest(c, "failed to put user on signup")
	} else if err = db.Save(NewPassword(model.UserID, creds.Password)); err != nil {
		errRequest(c, "failed to put password on signup")
	} else if err = db.Save(NewEmail(model.UserID, creds.Username)); err != nil {
		errRequest(c, "failed to put email on signup")
	}
	return model
}

func ResetPassword(c *gin.Context) {

	email, err := db.Find[Email](Data{"Address": creds(c).Username})
	if err != nil || email.UserID == ulid.Zero {
		badRequest(c, "failed to find email; try signing up?")
		return
	}

	tkn := NewResetPasswordToken(email.UserID)
	if err = db.Save(tkn); err != nil {
		errRequest(c, "failed to save forgot password token")
	} else if err = ses.SendPasswordReset(creds(c).Username, tkn.UserID.String()); err != nil {
		errRequest(c, "failed to send email on forgot password")
	} else {
		c.Status(http.StatusOK)
	}
}
