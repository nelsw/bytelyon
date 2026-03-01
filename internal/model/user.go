package model

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

var (
	tokenErr = errors.New("invalid JWT token (either expired or unprocessable")
)

type User struct {
	ID uuid.UUID `json:"id" dynamodbav:"ID,binary"`
}

func (u *User) Desc() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		BillingMode: types.BillingModeProvisioned,
		KeySchema: []types.KeySchemaElement{{
			AttributeName: Ptr("ID"),
			KeyType:       types.KeyTypeHash,
		}},
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: Ptr("ID"),
			AttributeType: types.ScalarAttributeTypeB,
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  Ptr(int64(10)),
			WriteCapacityUnits: Ptr(int64(10)),
		},
		TableName: Ptr(u.Name()),
	}
}

func (u *User) Name() string {
	return "ByteLyon_" + ModeTitle() + "_User"
}

func (u *User) Key() map[string]any {
	return map[string]any{"ID": u.ID}
}

func (u *User) Validate() error {
	return nil
}

func (u *User) NewJWT() (string, error) {

	log.Trace().Msg("creating JWT token")

	tkn, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
		ID:        uuid.NewString(),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Issuer:    "https://ByteLyon.com",
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		Subject:   u.ID.String(),
	}).SignedString(JwtKey())

	if err != nil {
		log.Err(err).Msg("error creating JWT token")
		return "", err
	}

	log.Debug().Msg("created JWT token")
	return tkn, nil
}

func NewUser(str string) (*User, error) {

	log.Trace().Msg("creating new user from JWT")

	tkn, err := jwt.ParseWithClaims(str, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) {
		return JwtKey(), nil
	})

	if err != nil {
		log.Err(err).Msg("failed to create new user (JWT parse err)")
		return nil, err
	}

	if !tkn.Valid {
		log.Warn().Msg("unable to create new user (JWT token invalid)")
		return nil, tokenErr
	}

	var id uuid.UUID
	if id, err = uuid.Parse(str); err != nil {
		log.Err(err).Msg("unable to create new user (UUID parse err)")
		return nil, err
	}

	log.Debug().Stringer("ID", id).Msg("created new user ")

	return &User{ID: id}, nil
}
