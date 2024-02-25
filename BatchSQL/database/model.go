package database

import "gorm.io/gorm"

type Tweet struct {
	gorm.Model

	Content string
}
