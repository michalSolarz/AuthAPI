package model

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

type (
	User struct {
		gorm.Model
		Email                string `json:"email" validate:"required,email"`
		Username             string `json:"username" validate:"required"`
		Password             string `json:"password" validate:"required,min=6"`
		PasswordConfirmation string `json:"passwordConfirmation" validate:"required,eqfield=Password"`
		UUID                 string `json:"uuid" validate:"required,uuid4"`
		Activated            bool `json:"activated"`
	}

	Password struct {
		Password             string `json:"password" validate:"required,min=6"`
		PasswordConfirmation string `json:"passwordConfirmation" validate:"required,eqfield=Password"`
	}

	CustomValidator struct {
		Validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}
