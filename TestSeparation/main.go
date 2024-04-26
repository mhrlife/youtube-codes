package main

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {

	dsc := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsc), &gorm.Config{})

	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
}
