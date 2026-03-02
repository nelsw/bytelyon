package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/nelsw/bytelyon/internal/service/ses"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

func Token(c *gin.Context) {

	ID, err := uuid.Parse(c.Param("token"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	var tkn model.Token
	if tkn, err = db.Find[model.Token](map[string]any{"ID": ID}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if tkn.ID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "token not found"})
		return
	}

	if tkn.Expiry.After(time.Now()) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "token expired"})
		return
	}

	if tkn.Type == model.ConfirmEmailTokenType {

		var jwt string
		if jwt, err = model.NewJWT(tkn.UserID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if err = db.Wipe(model.Token{}, map[string]any{"ID": ID}); err != nil {
			log.Warn().Err(err).Msg("failed to delete token")
		}

		c.JSON(http.StatusOK, gin.H{
			"isAuthenticated": true,
			"context": map[string]any{
				"token": jwt,
			},
		})
		return
	}

	if tkn.Type == model.ResetPasswordTokenType {

		var pass model.Password
		if pass, err = db.Find[model.Password](map[string]any{"ID": tkn.UserID}); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to find user and password from token"})
			return
		}

		a, ok := c.Get("creds")
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		pass.Update(a.(*model.Credentials).Password)
		if err = db.Save(&pass); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var jwt string
		if jwt, err = model.NewJWT(tkn.UserID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if err = db.Wipe(model.Token{}, map[string]any{"ID": ID}); err != nil {
			log.Warn().Err(err).Msg("failed to delete token")
		}

		c.JSON(http.StatusOK, gin.H{
			"isAuthenticated": true,
			"context": map[string]any{
				"token": jwt,
			},
		})
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unknown token type"})
}

func Login(c *gin.Context) {
	jwt, _ := c.Get("JWT")
	c.JSON(200, gin.H{
		"isAuthenticated": true,
		"context": map[string]any{
			"token": jwt,
		},
	})
}

func Signup(c *gin.Context) {
	a, ok := c.Get("creds")
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	creds := a.(*model.Credentials)

	// have we seen this email address before
	email, err := db.Find[model.Email](map[string]any{"ID": creds.Username})
	if err != nil {
		log.Err(err).Msg("failed to get email on signup")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// define a user here because we'll either put or get one below
	var user model.User

	// this email is new; create user & password models for finalizing the signup process
	if email.UserID == uuid.Nil {

		// create a user
		user = model.User{util.Must(uuid.NewV7())}
		if err = db.Save(user); err != nil {
			log.Err(err).Msg("failed to put user on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// create a password
		if err = db.Save(model.NewPassword(user.ID, creds.Password)); err != nil {
			log.Err(err).Msg("failed to put password on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// create email
		if err = db.Save(model.NewEmail(user.ID, creds.Username)); err != nil {
			log.Err(err).Msg("failed to put email on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	} else {
		// we've seen this email before
		if user, err = db.Find[model.User](map[string]any{"ID": email.UserID}); err != nil {
			log.Err(err).Msg("failed to get user on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
	}

	tkn := model.NewConfirmEmailToken(email.UserID)
	if err = db.Save(tkn); err != nil {
		log.Err(err).Msg("failed to put token on signup")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}

	if err = ses.SendEmailConfirmation(creds.Username, tkn.ID.String()); err != nil {
		log.Err(err).Msg("failed to send email on signup")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var jwt string
	if jwt, err = model.NewJWT(user.ID); err != nil {
		log.Err(err).Msg("failed to generate JWT token")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(200, gin.H{
		"isAuthenticated": true,
		"context": map[string]any{
			"token": jwt,
		},
	})
}

func ForgotPassword(c *gin.Context) {
	a, ok := c.Get("creds")
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	creds := a.(*model.Credentials)

	email, err := db.Find[model.Email](map[string]any{"ID": creds.Username})
	if err != nil {
		log.Err(err).Msg("failed to get email on forgot password")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// does homeboy exist? literally? not in a meta way.
	if email.UserID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "email not found; try signing up."})
		return
	}

	tkn := model.NewResetPasswordToken(email.UserID)
	if err = db.Save(tkn); err != nil {
		log.Err(err).Msg("failed to put token on forgot password")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}

	if err = ses.SendPasswordReset(creds.Username, tkn.ID.String()); err != nil {
		log.Err(err).Msg("failed to send email on forgot password")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

func ValidateAuth(c *gin.Context) {

	if c.Request.Method == http.MethodOptions {
		c.Next()
		return
	}

	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = c.Request.Header.Get("authorization")
	}

	tokenType, tokenValue, _ := strings.Cut(authHeader, " ")
	log.Info().Msgf("Token type: %s", tokenType)

	if tokenType == "Bearer" {
		validateBearerAuth(c, tokenValue)
		c.Next()
		return
	}

	if tokenType == "Basic" {
		validateBasicAuth(c, tokenValue)
		c.Next()
		return
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid authorization token; must be 'Bearer <token>' or 'Basic <token>'"})
}

func validateBasicAuth(c *gin.Context, token string) {

	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Error().Msgf("Error decoding basic auth token: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, password, ok := strings.Cut(string(b), ":")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid basic token; must be base64 encoded '<email>:<password>'"})
		return
	}

	creds := model.NewCredentials(username, password)

	if strings.Contains(c.FullPath(), "/token") {
		c.Set("creds", creds)
		c.Next()
		return
	}

	if err = creds.ValidateUsername(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
	}

	if strings.Contains(c.FullPath(), "/forgot-password") {
		c.Set("creds", creds)
		c.Next()
		return
	}

	if err = creds.ValidatePassword(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
	}

	if strings.Contains(c.FullPath(), "/signup") {
		c.Set("creds", creds)
		c.Next()
		return
	}

	var email model.Email
	if email, err = db.Find[model.Email](map[string]any{"ID": username}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("user not found for email %s", username),
		})
		return
	}

	if email.UserID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "email not found; try signing up."})
		return
	}

	var pass model.Password
	if pass, err = db.Find[model.Password](map[string]any{"ID": email.UserID}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if err = pass.Authenticate(password); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "password is incorrect"})
		return
	}

	var user model.User
	if user, err = db.Find[model.User](map[string]any{"ID": email.UserID}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var tkn string
	if tkn, err = model.NewJWT(user.ID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.Set("JWT", tkn)
}

func validateBearerAuth(c *gin.Context, token string) {

	log.Trace().Msg("validate bearer token")

	userID, err := model.ParseUserID(token)
	if err != nil {
		log.Warn().Err(err).Msg("invalid bearer token")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	log.Debug().Msg("bearer token validated")
	c.Set("userID", userID)
}
