package database

import (
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var _ TweetRepository = &SimpleRepository{}

type SimpleRepository struct {
	db    *gorm.DB
	group singleflight.Group
}

func NewSimpleRepository(db *gorm.DB) *SimpleRepository {
	return &SimpleRepository{db: db}
}

func (s *SimpleRepository) ByID(ctx context.Context, id uint) (Tweet, error) {
	var tweet Tweet
	if err := s.db.WithContext(ctx).Where("id=?", id).First(&tweet).Error; err != nil {
		logrus.WithError(err).Errorln("error while fetching tweet by id")
		return Tweet{}, err
	}
	return tweet, nil
}
