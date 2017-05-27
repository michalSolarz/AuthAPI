package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/labstack/echo/middleware"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"./handler"
	"./model"
	"gopkg.in/go-playground/validator.v9"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
	"github.com/go-redis/redis"
	"github.com/adjust/redismq"
)

const (
	DB_NAME        = "redikop"
	DB_USER        = "redikop"
	DB_PASSWORD    = "redikop"
	REDIS_HOST     = "localhost"
	REDIS_PORT     = "6379"
	REDIS_PASSWORD = ""
)

func main() {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := gorm.Open("postgres", connectionString)
	checkError(err)
	defer db.Close()
	db.AutoMigrate(&model.User{})

	tokenStorage := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", REDIS_HOST, REDIS_PORT), Password: REDIS_PASSWORD, DB: 0})
	redisError := tokenStorage.Ping().Err()
	checkError(redisError)

	redisConnections := map[string]*redis.Client{
		"tokenStorage": tokenStorage}


	mailingQueue := redismq.CreateQueue(REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, 10, "mailingQueue")

	app := echo.New()
	app.Logger.SetLevel(log.DEBUG)
	validate := validator.New()
	app.Validator = &model.CustomValidator{Validator: validate}
	en := en.New()
	uni := ut.New(en, en)
	translation, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, translation)

	app.Use(middleware.Logger())

	h := &handler.Handler{DB: db,
		Translation:      translation,
		Config:           map[string]string{"secret": "VerySecretSecret"},
		RedisConnections: redisConnections,
		MailingQueue:     mailingQueue,
	}

	app.POST("/sign-up", h.SignUp)
	app.GET("/activate-account/:userUuid/:token", h.ActivateAccount)
	app.POST("/login", h.Login)
	app.POST("/request-password-reset", h.RequestPasswordReset)
	app.GET("/password-reset/:userUuid/:token", h.PasswordResetAttempt)
	app.POST("/password-reset", h.PasswordReset)

	app.Logger.Fatal(app.Start(":8080"))
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
