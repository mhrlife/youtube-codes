package database

import "context"

type TweetRepository interface {
	ByID(ctx context.Context, id uint) (Tweet, error)
}
