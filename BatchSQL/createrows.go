package main

import (
	"BatchSQL/database"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"strconv"
)

func main() {
	godotenv.Load()
	db, err := database.NewConnection()
	if err != nil {
		logrus.WithError(err).Errorln("couldn't connect to the database")
		return
	}

	if err := db.AutoMigrate(&database.Tweet{}); err != nil {
		logrus.WithError(err).Errorln("error while migrating the database")
		return
	}

	// creating tweets in batches
	var tweets []database.Tweet
	for i := 0; i < 100_000; i++ {
		tweets = append(tweets, database.Tweet{Content: "A Random Tweet #" + strconv.Itoa(i)})
	}

	if err := db.CreateInBatches(tweets, 1000).Error; err != nil {
		logrus.WithError(err).Errorln("error while creating users")
		return
	}

	logrus.Println("users created successfully!")
}
