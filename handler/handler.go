package handler

import (
	"github.com/jinzhu/gorm"
	"github.com/go-playground/universal-translator"
)

type (
	Handler struct {
		DB *gorm.DB
		Translation ut.Translator
		Config map[string]string
	}
)
