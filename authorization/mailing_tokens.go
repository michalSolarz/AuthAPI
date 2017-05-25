package authorization

import (
	"gitlab.com/michalSolarz/AuthAPI/model"
	"github.com/satori/go.uuid"
	"fmt"
	"github.com/go-redis/redis"
	"time"
	"github.com/labstack/gommon/log"
)

type (
	MailingToken struct {
		Email     string `json:"email"`
		Token     string `json:"token"`
		UserUuid  string `json:"user_uuid"`
		TokenType string `json:"token_type"`
		Expires   string `json:"expires"`
	}
)

const (
	AccountActivationTokenType = "AccountActivation"
	PasswordResetTokenType     = "PasswordReset"
)

func NewMailingToken(user *model.User, tokenType string, expires *time.Time) MailingToken {
	return MailingToken{Token: uuid.NewV4().String(), Email: user.Email, UserUuid: user.UUID, TokenType: tokenType, Expires: expires.Format("2006-01-02 15:04:05.999999999")}
}

func MailingTokenToRedis(redisConnection *redis.Client, token *MailingToken) {
	redisConnection.HSet(fmt.Sprintf("User:%s:MailingToken:%s", token.UserUuid, token.TokenType), token.Token, token.Expires)
}

func MailingTokenExpiration(tokenType string) *time.Time {
	currentTime := time.Now().UTC()
	var expiration time.Time

	if tokenType == AccountActivationTokenType {
		expiration = currentTime.AddDate(0, 0, 1)
	} else if tokenType == PasswordResetTokenType {
		expiration = currentTime.Add(time.Hour * time.Duration(1))
	} else {
		expiration = currentTime.Add(time.Minute * time.Duration(30))
	}
	return &expiration
}

func MailingTokenInRedis(redisConnection *redis.Client, token *MailingToken) bool {
	cmd := redisConnection.HGet(fmt.Sprintf("User:%s:MailingToken:%s", token.UserUuid, token.TokenType), token.Token)
	t, err := cmd.Result()
	if err == redis.Nil {
		return false
	} else if err != nil {
		log.Error(err)
		return false
	}

	expired, err := time.Parse("2006-01-02 15:04:05.999999999", t)
	if err != nil {
		log.Error(err)
		return false
	}

	isExpired := expired.Sub(time.Now().UTC()) <= 0
	if isExpired {
		InvalidateMailingToken(redisConnection, token)
		return false
	}

	return true
}

func InvalidateMailingToken(redisConnection *redis.Client, token *MailingToken) {
	redisConnection.HDel(fmt.Sprintf("User:%s:MailingToken:%s", token.UserUuid, token.TokenType), token.Token)
}
