package authorization

import (
	"gitlab.com/michalSolarz/AuthAPI/model"
	"github.com/satori/go.uuid"
	"fmt"
	"github.com/go-redis/redis"
)

type (
	MailingToken struct {
		Email     string `json:"email"`
		Token     string `json:"token"`
		UserUuid  string `json:"user_uuid"`
		TokenType string `json:"token_type"`
	}
)

const (
	AccountActivationTokenType = "AccountActivation"
	PasswordResetTokenType     = "PasswordReset"
)

func NewMailingToken(user *model.User, tokenType string) MailingToken {
	return MailingToken{Token: uuid.NewV4().String(), Email: user.Email, UserUuid: user.UUID, TokenType: tokenType}
}

func MailingTokenToRedis(redisConnection *redis.Client, token *MailingToken) {
	redisConnection.SAdd(fmt.Sprintf("User:%s:MailingToken:%s", token.UserUuid, token.TokenType), token.Token)
}
