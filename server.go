package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/labstack/echo/middleware"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
)

const (
	DB_NAME     = "redikop"
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
)

func main() {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", connectionString)
	checkError(err)
	defer db.Close()

	app := echo.New()
	app.Logger.SetLevel(log.DEBUG)
	app.Use(middleware.Logger())

	//h := &handler.Handler{DB: db}

	//app.POST("/sign-up", h.SignUp)


	app.Logger.Fatal(app.Start(":8080"))
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
