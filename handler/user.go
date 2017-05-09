package handler

import (
	"github.com/labstack/echo"
	"net/http"
	"gitlab.com/michalSolarz/AuthAPI/model"
	"gopkg.in/go-playground/validator.v9"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
	"github.com/go-redis/redis"
)

func (h *Handler) SignUp(c echo.Context) (err error) {
	u := &model.User{}
	if err = c.Bind(u); err != nil {
		return
	}

	u.UUID = uuid.NewV4().String()

	if err = c.Validate(u); err != nil {
		errs := err.(validator.ValidationErrors)
		return c.JSON(http.StatusUnprocessableEntity, errs.Translate(h.Translation))
	}

	existingUsers := []model.User{}
	h.DB.Where("email LIKE ? OR username LIKE ?", u.Email, u.Username).Find(&existingUsers)
	if len(existingUsers) != 0 {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": fmt.Sprintf("User with email: %s or username: %s already exists", u.Email, u.Username)})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		c.Logger().Error(fmt.Sprintf("Failed to hash password: %s", u.Password))
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to hash password"})
	}
	u.Password = string(hashedPassword)

	h.DB.Create(u)

	tokenId, token, err := GenerateToken([]byte(h.Config["secret"]))
	TokenIdToRedis(h.RedisConnections["tokenStorage"], tokenId)
	if err != nil {
		c.Logger().Error("Failed to generate token")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to generate token"})
	}
	c.Response().Header().Add("auth-token", string(token))

	return c.JSON(http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) Login(c echo.Context) (err error) {
	u := &model.User{}
	if err = c.Bind(u); err != nil {
		c.Logger().Error("Failed to bind existingUser data")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to bind existingUser data"})
	}

	existingUser := []model.User{}
	h.DB.Where("username LIKE ?", u.Username).Find(&existingUser)
	if len(existingUser) == 0 {
		bcrypt.GenerateFromPassword(uuid.NewV4().Bytes(), 12)
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Missing user or invalid password"})
	}

	isValidPassword := bcrypt.CompareHashAndPassword([]byte(existingUser[0].Password), []byte(u.Password))

	if isValidPassword != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Missing user or invalid password"})
	}

	tokenId, token, err := GenerateToken([]byte(h.Config["secret"]))
	TokenIdToRedis(h.RedisConnections["tokenStorage"], tokenId)
	if err != nil {
		c.Logger().Error("Failed to generate token")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to generate token"})
	}

	c.Response().Header().Add("auth-token", string(token))

	return c.JSON(http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) ResetPassword(c echo.Context) (err error) {
	return c.JSON(http.StatusCreated, map[string]string{"hello": "reset-password"})
}

func (h *Handler) LoginFacebook(c echo.Context) (err error) {
	return c.JSON(http.StatusCreated, map[string]string{"hello": "login-facebook"})
}

func (h *Handler) LoginGoogle(c echo.Context) (err error) {
	return c.JSON(http.StatusCreated, map[string]string{"hello": "login-google"})
}

func GenerateToken(signKey []byte) (tokenId string, tokenString string, err error) {
	id := uuid.NewV4().String()
	claims := &jwt.StandardClaims{Id: id, NotBefore: time.Now().Unix(), IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix()}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(signKey)

	return id, signedString, err
}

func TokenIdToRedis(redisConnection *redis.Client, tokenId string) {
	redisConnection.SAdd("activeTokensStore", tokenId)
	TokenIdToDailyStorage(redisConnection, tokenId)
}

func TokenIdToDailyStorage(redisConnection *redis.Client, tokenId string) {
	date := time.Now().UTC().Format("2006-01-02")
	redisConnection.SAdd("dailyTokensStore:"+date, tokenId)
	AddDailyStorage(redisConnection, date)
}

func AddDailyStorage(redisConnection *redis.Client, date string) {
	redisConnection.SAdd("dailyTokensStores", date)
}
