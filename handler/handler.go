package handler

import (
	"github.com/jinzhu/gorm"
	"github.com/go-playground/universal-translator"
	"github.com/go-redis/redis"
	"github.com/adjust/redismq"
)

type (
	Handler struct {
		DB               *gorm.DB
		Translation      ut.Translator
		Config           map[string]string
		RedisConnections map[string]*redis.Client
		MailingQueue     *redismq.Queue
	}
)
