package database

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func NewGorm(masterDSN string, replicaDSNs ...string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open("test.db"), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Error("couldn't connect to the database")
		return nil, err
	}

	if err := db.Use(dbresolver.Register(dbresolver.Config{
		Replicas: lo.Map(append(replicaDSNs, masterDSN), func(item string, _ int) gorm.Dialector {
			return mysql.Open(item)
		}),
	})); err != nil {
		logrus.WithError(err).Error("couldn't setup replica databases")
		return nil, err
	}

	return db, nil
}
