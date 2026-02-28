package handler

import (
	"encoding/base64"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	client "github.com/nelsw/bytelyon/internal/client/dynamodb"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
)

var (
	urlValidationRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
)

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

	tokenType, tokenValue, _ := strings.Cut(authHeader, " ")
	if tokenType == "Bearer" {
		validateBearerAuth(c, tokenValue)
		return
	}

	if tokenType == "Basic" {
		validateBasicAuth(c, tokenValue)
		return
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid authorization token; must be 'Bearer <token>' or 'Basic <token>'"})
}

type Claims struct {
	*model.User `json:"data"`
	jwt.RegisteredClaims
}

func validateBasicAuth(c *gin.Context, token string) {

	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, password, ok := strings.Cut(string(b), ":")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid basic token; must be base64 encoded '<email>:<password>'"})
		return
	}

	dbc := client.New()

	var email model.Email
	if email, err = client.GetItem[model.Email](c, dbc, "Email", model.Email{ID: username}); err != nil {
		return
	}

	var pass model.Password
	if pass, err = client.GetItem[model.Password](c, dbc, "Password", email); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err = pass.Authenticate(password); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if user, err = client.GetItem[model.User](c, dbc, "User", email); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var tkn string
	if tkn, err = user.NewJWT(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tkn})
}

func validateBearerAuth(c *gin.Context, token string) {
	usr, err := model.NewUser(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Set("user", usr)
}
