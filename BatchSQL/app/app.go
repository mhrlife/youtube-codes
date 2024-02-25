package app

import (
	"BatchSQL/database"
	"context"
	"math/rand"
)

type App struct {
	TweetRepository database.TweetRepository
}

func NewApp(tweetRepository database.TweetRepository) *App {
	return &App{TweetRepository: tweetRepository}
}

func (a *App) RandomTweet(ctx context.Context) (database.Tweet, error) {
	id := rand.Intn(100_000) + 1
	return a.TweetRepository.ByID(ctx, uint(id))
}
