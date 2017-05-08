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
	h.DB.Where("email LIKE ? OR username LIKE ?", "Hello@world.pl", "hello123").Find(&existingUsers)
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

	token, err := GenerateToken([]byte(h.Config["secret"]))
	if err != nil {
		c.Logger().Error("Failed to generate token")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to generate token"})
	}
	c.Response().Header().Add("auth-token", string(token))

	return c.JSON(http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) Login(c echo.Context) (err error) {
	return c.JSON(http.StatusCreated, map[string]string{"hello": "login"})
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

func GenerateToken(signKey []byte) (string, error) {
	id := uuid.NewV4().String()
	claims := &jwt.StandardClaims{Id: id, NotBefore: time.Now().Unix(), IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix()}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signKey)

	return tokenString, err
}
