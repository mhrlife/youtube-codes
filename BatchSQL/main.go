package main

import (
	"BatchSQL/app"
	"BatchSQL/database"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func main() {
	// setup app
	godotenv.Load()
	db, err := database.NewConnection()
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't init the database")
	}
	appWithBatch := app.NewApp(database.NewBatchRepository(db, 1000))
	appWithSimple := app.NewApp(database.NewSimpleRepository(db))

	// setup router
	e := echo.New()
	e.GET("/random-tweet-simple", func(c echo.Context) error {
		tweet, err := appWithSimple.RandomTweet(c.Request().Context())
		if err != nil {
			return c.JSON(500, err)
		}
		return c.JSON(200, tweet)
	})

	e.GET("/random-tweet-batch", func(c echo.Context) error {
		tweet, err := appWithBatch.RandomTweet(c.Request().Context())
		if err != nil {
			return c.JSON(500, err)
		}
		return c.JSON(200, tweet)
	})

	e.Logger.Error(e.Start(":8080"))
}
