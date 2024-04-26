package repository

import (
	"TestSeparation/internal/entity"
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("user not found")
)

//go:generate mockery --name UsersRepository
type UsersRepository interface {
	ByField(ctx context.Context, field string, value any) (entity.User, error)
	Create(ctx context.Context, user *entity.User) error
}

type gormUsersRepository struct {
	db *gorm.DB
}

func NewGormUsersRepository(db *gorm.DB) UsersRepository {
	return &gormUsersRepository{db}
}

func (r *gormUsersRepository) ByField(ctx context.Context, field string, value interface{}) (entity.User, error) {
	var user entity.User
	if err := r.db.Where(field+" = ?", value).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.User{}, ErrNotFound
		}

		logrus.WithError(err).WithFields(logrus.Fields{

			"field": field,
			"value": value,
		}).Errorln("Error while searching for user by field")

		return entity.User{}, err
	}
	return user, nil
}

func (r *gormUsersRepository) Create(ctx context.Context, user *entity.User) error {
	err := r.db.Create(user).Error

	if err != nil {
		logrus.WithError(err).WithField("user", user).Error("Error while creating user")
		return err
	}

	return nil
}
