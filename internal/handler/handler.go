package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	dbClient "github.com/nelsw/bytelyon/internal/client/dynamodb"
	seClient "github.com/nelsw/bytelyon/internal/client/ses"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

var (
	urlValidationRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
)

func Token(c *gin.Context) {
	ID, err := uuid.Parse(c.Param("token"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dbc *dynamodb.Client
	if dbc, err = dbClient.New(); err != nil {
		panic(err)
	}

	var tkn model.Token
	if tkn, err = dbClient.GetItem[model.Token](c, dbc, &model.Token{ID: ID}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		// todo - func or prop to denote this email is confirmed
	} else if tkn.Type == model.ResetPasswordTokenType {

		var pass model.Password
		if pass, err = dbClient.GetItem[model.Password](c, dbc, &model.Password{ID: tkn.UserID}); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to find user and password from token"})
			return
		}

		a, ok := c.Get("creds")
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		pass.Update(a.(*model.Credentials).Password)
		if err = dbClient.PutItem(c, dbc, &pass); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err = dbClient.DeleteItem[model.Token](c, dbc, &model.Token{ID: ID}); err != nil {
		log.Warn().Err(err).Msg("failed to delete token")
	}

	c.Status(http.StatusOK)
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

	dbc, err := dbClient.New()
	if err != nil {
		panic(err)
	}

	// have we seen this email address before
	var email model.Email
	if email, err = dbClient.GetItem[model.Email](c, dbc, &model.Email{ID: creds.Username}); err != nil {
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
		if err = dbClient.PutItem(c, dbc, &user); err != nil {
			log.Err(err).Msg("failed to put user on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// create a password
		if err = dbClient.PutItem(c, dbc, model.NewPassword(&user, creds.Password)); err != nil {
			log.Err(err).Msg("failed to put password on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// create email
		if err = dbClient.PutItem(c, dbc, model.NewEmail(&user, creds.Username)); err != nil {
			log.Err(err).Msg("failed to put email on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	} else {
		// we've seen this email before
		if user, err = dbClient.GetItem[model.User](c, dbc, &model.User{ID: email.UserID}); err != nil {
			log.Err(err).Msg("failed to get user on signup")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tkn := model.NewConfirmEmailToken(email.UserID)
	if err = dbClient.PutItem(c, dbc, tkn); err != nil {
		log.Err(err).Msg("failed to put token on signup")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}

	sec := util.Must(seClient.New())
	if err = seClient.SendEmailConfirmation(c, sec, creds.Username, tkn.ID.String()); err != nil {
		log.Err(err).Msg("failed to send email on signup")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var jwt string
	if jwt, err = user.NewJWT(); err != nil {
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

	dbc, err := dbClient.New()
	if err != nil {
		panic(err)
	}

	var email model.Email
	if email, err = dbClient.GetItem[model.Email](c, dbc, &model.Email{ID: creds.Username}); err != nil {
		log.Err(err).Msg("failed to get email on forgot password")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	tkn := model.NewResetPasswordToken(email.UserID)
	if err = dbClient.PutItem(c, dbc, tkn); err != nil {
		log.Err(err).Msg("failed to put token on forgot password")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}

	sec := util.Must(seClient.New())
	if err = seClient.SendPasswordReset(c, sec, email.ID, tkn.ID.String()); err != nil {
		log.Err(err).Msg("failed to send email on forgot password")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

func Delete[T any](c *gin.Context) {
	if _, err := db.Builder[T]().Where("id = ?", c.MustGet("ID").(uint)).Delete(c); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func ListSearches(c *gin.Context) {

	arr, err := db.Builder[model.Search]().
		Preload("Bot", nil).
		Preload("Pages", nil).
		Where("bot_id = ?", c.MustGet("ID").(uint)).
		Order("created_at desc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func ListSitemaps(c *gin.Context) {

	arr, err := db.Builder[model.Sitemap]().
		Preload("Bot", nil).
		Where("bot_id = ?", c.MustGet("ID").(uint)).
		Order("created_at desc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func ListNews(c *gin.Context) {

	arr, err := db.Builder[model.News]().
		Where("bot_id = ?", c.MustGet("ID").(uint)).
		Order("published desc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func ListBots(c *gin.Context) {

	arr, err := db.Builder[model.Bot]().
		Order("target asc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func ListBotsByType(c *gin.Context) {
	t := model.BotType(c.Param("type"))
	if err := t.Validate(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	arr, err := db.Builder[model.Bot]().
		Where("type = ?", t).
		Order("target asc").
		Find(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arr)
}

func CreateBot(c *gin.Context) {
	var bot model.Bot
	if err := c.Bind(&bot); err != nil {
		return
	}
	if bot.Type == model.SitemapBotType {
		if ok := urlValidationRegex.MatchString(bot.Target); !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad url, must begin with https://"})
			return
		}
	}

	if err := db.Builder[model.Bot]().Create(c, &bot); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bot)
}

func UpdateBot(c *gin.Context) {
	var bot model.Bot
	if err := c.Bind(&bot); err != nil {
		return
	}
	if bot.Type == model.SitemapBotType {
		if ok := urlValidationRegex.MatchString(bot.Target); !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad url, must begin with https://"})
			return
		}
	}
	_, err := db.Builder[model.Bot]().
		Where("id = ?", bot.ID).
		Updates(c, bot)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bot)
}

func ValidateID(c *gin.Context) {

	if !strings.Contains(c.FullPath(), ":id") {
		c.Next()
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Set("ID", uint(id))
	c.Next()
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

	log.Info().Msgf("Authorization header: %s", authHeader)

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
	if err = creds.ValidateUsername(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else if err = creds.ValidatePassword(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if strings.Contains(c.FullPath(), "/signup") ||
		strings.Contains(c.FullPath(), "/token") ||
		strings.Contains(c.FullPath(), "/forgot-password") {
		c.Set("creds", creds)
		c.Next()
		return
	}

	dbc := util.Must(dbClient.New())

	var email model.Email
	if email, err = dbClient.GetItem[model.Email](c, dbc, &model.Email{ID: username}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("user not found for email %s", username),
		})
		return
	}

	var pass model.Password
	if pass, err = dbClient.GetItem[model.Password](c, dbc, &model.Password{ID: email.UserID}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err = pass.Authenticate(password); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "password is incorrect"})
		return
	}

	var user model.User
	if user, err = dbClient.GetItem[model.User](c, dbc, &model.User{ID: email.UserID}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var tkn string
	if tkn, err = user.NewJWT(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("JWT", tkn)
}

func validateBearerAuth(c *gin.Context, token string) {

	log.Trace().
		Str("token", token).
		Msg("validate bearer token")

	usr, err := model.NewUser(token)
	if err != nil {
		log.Warn().Err(err).Msg("invalid bearer token")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Debug().Msg("bearer token validated")
	c.Set("user", usr)
}
