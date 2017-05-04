package handler

import (
	"github.com/labstack/echo"
	"net/http"
)

func (h *Handler) SignUp(c echo.Context) (err error) {

	return c.JSON(http.StatusCreated, map[string]string{"hello": "world"})
}
