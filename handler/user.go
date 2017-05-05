package handler

import (
	"github.com/labstack/echo"
	"net/http"
	"gitlab.com/michalSolarz/AuthAPI/model"
	"gopkg.in/go-playground/validator.v9"
	"github.com/satori/go.uuid"
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

	h.DB.Create(u)

	return c.JSON(http.StatusCreated, map[string]string{"hello": "world"})
}
