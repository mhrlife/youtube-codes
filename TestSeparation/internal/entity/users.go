package entity

import "gorm.io/gorm"

type User struct {
	gorm.Model

	PhoneNumber string
	DisplayName string
}
