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
	"gitlab.com/michalSolarz/AuthAPI/mailing_queue"
)

func (h *Handler) SignUp(c echo.Context) (err error) {
	u := &model.User{Activated: false}
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

	authorizationToken, err := authorization.GenerateToken([]byte(h.Config["secret"]), c, u.UUID)
	authorization.TokenToRedis(h.RedisConnections["tokenStorage"], authorizationToken)

	activationToken := authorization.NewMailingToken(u, authorization.AccountActivationTokenType, authorization.MailingTokenExpiration(authorization.AccountActivationTokenType))

	authorization.MailingTokenToRedis(h.RedisConnections["tokenStorage"], &activationToken)
	mailing_queue.QueueTransactionalMail(h.MailingQueue, activationToken)

	if err != nil {
		c.Logger().Error("Failed to generate authorizationToken")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to generate authorizationToken"})
	}
	c.Response().Header().Add("auth-authorizationToken", authorizationToken.SignedString)

	return c.JSON(http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) ActivateAccount(c echo.Context) (err error) {
	mailingToken := authorization.MailingToken{Token: c.Param("token"), UserUuid: c.Param("userUuid"), TokenType: authorization.AccountActivationTokenType}
	inRedis := authorization.MailingTokenInRedis(h.RedisConnections["tokenStorage"], &mailingToken)
	if !inRedis {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}

	user := []model.User{}
	h.DB.Where("uuid LIKE ?", mailingToken.UserUuid).Find(&user)
	if len(user) == 0 || user[0].Activated == true {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}

	h.DB.Model(&user).Update("activated", true)
	authorization.InvalidateMailingToken(h.RedisConnections["tokenStorage"], &mailingToken)

	return c.JSON(http.StatusOK, map[string]string{"hello": fmt.Sprintf("ActivateAccount UserUUID:%s", c.Param("userUuid"))})
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

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) RequestPasswordReset(c echo.Context) (err error) {
	u := &model.User{}
	if err = c.Bind(u); err != nil {
		c.Logger().Error("Failed to bind existingUser data")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to bind existingUser data"})
	}

	existingUser := []model.User{}
	h.DB.Where("email LIKE ?", u.Email).Find(&existingUser)
	if len(existingUser) == 0 {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "User doesn't exist"})
	}

	passwordResetToken := authorization.NewMailingToken(&existingUser[0], authorization.PasswordResetTokenType, authorization.MailingTokenExpiration(authorization.PasswordResetTokenType))

	authorization.MailingTokenToRedis(h.RedisConnections["tokenStorage"], &passwordResetToken)
	mailing_queue.QueueTransactionalMail(h.MailingQueue, passwordResetToken)

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) PasswordResetAttempt(c echo.Context) (err error) {
	mailingToken := authorization.MailingToken{Token: c.Param("token"), UserUuid: c.Param("userUuid"), TokenType: authorization.PasswordResetTokenType}
	inRedis := authorization.MailingTokenInRedis(h.RedisConnections["tokenStorage"], &mailingToken)
	if !inRedis {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}

	user := []model.User{}
	h.DB.Where("uuid LIKE ?", mailingToken.UserUuid).Find(&user)
	if len(user) == 0 {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}

	passwordResetToken := uuid.NewV4().String()

	authorization.PasswordResetTokenToRedis(h.RedisConnections["tokenStorage"], passwordResetToken)
	authorization.InvalidateMailingToken(h.RedisConnections["tokenStorage"], &mailingToken)

	c.Response().Header().Add("auth-password-reset-id", mailingToken.UserUuid)
	c.Response().Header().Add("auth-password-reset-token", passwordResetToken)
	return c.JSON(http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) PasswordReset(c echo.Context) (err error) {
	token := c.Request().Header.Get("auth-password-reset-token")
	userUuid := c.Request().Header.Get("auth-password-reset-id")
	if len(userUuid) == 0 {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}

	inStore := authorization.PasswordResetTokenInRedis(h.RedisConnections["tokenStorage"], token)
	if !inStore {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}

	user := []model.User{}
	h.DB.Where("uuid LIKE ?", userUuid).Find(&user)
	if len(user) == 0 {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}

	p := &model.Password{}
	if err = c.Bind(p); err != nil {
		return
	}

	if err = c.Validate(p); err != nil {
		errs := err.(validator.ValidationErrors)
		return c.JSON(http.StatusUnprocessableEntity, errs.Translate(h.Translation))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), 12)
	if err != nil {
		c.Logger().Error(fmt.Sprintf("Failed to hash password: %s", p.Password))
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Failed to hash password"})
	}
	h.DB.Model(&user).Update("password", string(hashedPassword))
	authorization.RemovePasswordResetTokenFromRedis(h.RedisConnections["tokenStorage"], token)

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) LoginFacebook(c echo.Context) (err error) {
	return c.JSON(http.StatusCreated, map[string]string{"hello": "login-facebook"})
}

func (h *Handler) LoginGoogle(c echo.Context) (err error) {
	return c.JSON(http.StatusCreated, map[string]string{"hello": "login-google"})
}
