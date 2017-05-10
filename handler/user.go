package handler

import (
	"github.com/labstack/echo"
	"net/http"
	"gitlab.com/michalSolarz/AuthAPI/model"
	"gopkg.in/go-playground/validator.v9"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"gitlab.com/michalSolarz/AuthAPI/authorization"
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

	token, err := authorization.GenerateToken([]byte(h.Config["secret"]), c, u.UUID)
	authorization.TokenToRedis(h.RedisConnections["tokenStorage"], token)
	if err != nil {
		c.Logger().Error("Failed to generate token")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to generate token"})
	}
	c.Response().Header().Add("auth-token", token.SignedString)

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

	token, err := authorization.GenerateToken([]byte(h.Config["secret"]), c, existingUser[0].UUID)
	authorization.TokenToRedis(h.RedisConnections["tokenStorage"], token)
	if err != nil {
		c.Logger().Error("Failed to generate token")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to generate token"})
	}

	c.Response().Header().Add("auth-token", token.SignedString)

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