package main

import (
	"TestSeparation/internal/repository"
	"TestSeparation/internal/service"
	"context"

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

	usersRepository := repository.NewGormUsersRepository(db)
	usersService := service.NewUsersService(usersRepository)

	usersService.Signup(context.Background(), "09211231231", "Mohammad")
}
