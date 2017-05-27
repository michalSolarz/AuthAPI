package authorization

import (
	"github.com/satori/go.uuid"
	"github.com/dgrijalva/jwt-go"
	"time"
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"encoding/json"
	"fmt"
)

type Token struct {
	ID           string
	SignedString string
	Info         TokenInfo
	UserUuid     string
}

type TokenInfo struct {
	CreatedAt string
	IP        string
	UserAgent string
}

func GenerateToken(signKey []byte, c echo.Context, userUuid string) (token Token, err error) {
	issued := time.Now().UTC()
	tokenInfo := TokenInfo{issued.Format("2006-01-02 15:04:05.999999999"), c.RealIP(), c.Request().UserAgent()}
	token = Token{ID: uuid.NewV4().String(), Info: tokenInfo, UserUuid: userUuid}
	claims := &jwt.StandardClaims{Id: token.ID, NotBefore: issued.Unix(), IssuedAt: issued.Unix(), ExpiresAt: issued.Add(time.Hour * 24 * 30).Unix()}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := jwtToken.SignedString(signKey)
	token.SignedString = string(signedString)

	return token, err
}

func TokenToRedis(redisConnection *redis.Client, token Token) {
	redisConnection.SAdd("activeTokensStore", token.ID)
	TokenToDailyStorage(redisConnection, token)
	TokenToUserStorage(redisConnection, token)
}

func TokenToDailyStorage(redisConnection *redis.Client, token Token) {
	date := time.Now().UTC().Format("2006-01-02")
	redisConnection.SAdd("dailyTokensStore:"+date, token.ID)
	AddDailyStorage(redisConnection, date)
}

func TokenToUserStorage(redisConnection *redis.Client, token Token) {
	info, _ := json.Marshal(token.Info)
	redisConnection.HSet("userTokenStore:"+token.UserUuid, token.ID, string(info))
}

func AddDailyStorage(redisConnection *redis.Client, date string) {
	redisConnection.SAdd("dailyTokensStores", date)
}

func PasswordResetTokenToRedis(redisConnection *redis.Client, token string) {
	redisConnection.Set(fmt.Sprintf("passwordResetToken:%s", token), token, time.Duration(time.Minute*5))
}

func PasswordResetTokenInRedis(redisConnection *redis.Client, token string) bool {
	t := redisConnection.Get(fmt.Sprintf("passwordResetToken:%s", token))
	if t.Err() == redis.Nil {
		return false
	}
	return true
}

func RemovePasswordResetTokenFromRedis(redisConnection *redis.Client, token string)  {
	redisConnection.Del(fmt.Sprintf("passwordResetToken:%s", token))
}
