package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/logger"
	. "github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/rs/zerolog/log"
)

var errJSON = func(err any) Data {
	return Data{"error": err}
}

func botType(c *gin.Context) BotType    { return c.MustGet("BOT_TYPE").(BotType) }
func userID(c *gin.Context) uuid.UUID   { return c.MustGet("USER_ID").(uuid.UUID) }
func creds(c *gin.Context) *Credentials { return c.MustGet("CREDS").(*Credentials) }

var tokenResponse = func(c *gin.Context, a any) {
	c.JSON(http.StatusOK, Data{
		"isAuthenticated": true,
		"context": Data{
			"token": a,
		},
	})
}

var badRequest = func(c *gin.Context, a any) {
	log.Warn().Msgf("Bad Request: %v", a)
	c.AbortWithStatusJSON(http.StatusBadRequest, Data{"error": fmt.Sprint(a)})
}
var unauthorized = func(c *gin.Context, a any) {
	log.Warn().Msgf("Unauthorized: %v", a)
	c.AbortWithStatusJSON(http.StatusUnauthorized, Data{"error": fmt.Sprint(a)})
}
var errRequest = func(c *gin.Context, a any) {
	log.Error().Msgf("Err Request: %v", a)
	c.AbortWithStatusJSON(http.StatusInternalServerError, Data{"error": fmt.Sprint(a)})
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		fmt.Printf("%s GIN %s > %s %v\n",
			logger.BlackIntense+time.Now().Format("15:04:05")+logger.Default,
			logger.BlackBackground+c.FullPath()+logger.Cyan,
			logger.Default+time.Since(t).String(),
			c.Writer.Status(),
		)
	}
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
		if validateBasicAuth(c, tokenValue) {
			c.Next()
		}
		return
	}

	if strings.Contains(c.FullPath(), "/token") {
		c.Next()
		return
	}
	unauthorized(c, "invalid authorization token; must be 'Bearer <token>' or 'Basic <token>'")
}

func validateBasicAuth(c *gin.Context, token string) (ok bool) {

	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		badRequest(c, "Error decoding basic auth token")
		return
	}

	username, password, ok := strings.Cut(string(b), ":")
	if !ok && strings.Contains(c.FullPath(), "/token") {
		c.Next()
		return
	}

	if !ok {
		badRequest(c, "invalid basic token; must be base64 encoded '<email>:<password>'")
		return
	}

	credentials := NewCredentials(username, password)

	if strings.Contains(c.FullPath(), "/token") {
		c.Set("CREDS", credentials)
		c.Next()
		return
	}

	if err = credentials.ValidateUsername(); err != nil {
		badRequest(c, err)
		return
	}

	if strings.Contains(c.FullPath(), "/forgot-password") {
		c.Set("CREDS", credentials)
		c.Next()
		return
	}

	log.Trace().Msg("validating password")
	err = credentials.ValidatePassword()
	log.Debug().Err(err).Msg("validated password")

	if err != nil {
		unauthorized(c, err.Error())
		return
	}

	if strings.Contains(c.FullPath(), "signup") {
		c.Set("CREDS", credentials)
		c.Next()
		return
	}

	var email Email
	if email, err = db.Find[Email](Data{"Address": username}); err != nil {
		badRequest(c, err)
	} else if email.UserID == uuid.Nil {
		badRequest(c, "email not found; try signing up.")
	}

	var pass Password
	if pass, err = db.Find[Password](Data{"UserID": email.UserID}); err != nil {
		badRequest(c, err)
	} else if err = pass.Authenticate(password); err != nil {
		unauthorized(c, "password is incorrect")
	}

	var jwt string
	if jwt, err = NewJWT(email.UserID); err != nil {
		badRequest(c, err)
	}
	c.Set("JWT", jwt)
	return true
}

func validateBearerAuth(c *gin.Context, token string) {
	log.Trace().Msg("validate bearer token")
	if id, err := ParseUserID(token); err != nil {
		unauthorized(c, "invalid bearer token")
	} else {
		log.Debug().Msg("validated bearer token")
		c.Set("USER_ID", id)
	}
}

func ValidateBotType(c *gin.Context) {
	if typ, err := NewBotType(c.Param("type")); err != nil {
		badRequest(c, err)
	} else {
		c.Set("BOT_TYPE", typ)
		c.Next()
	}
}
