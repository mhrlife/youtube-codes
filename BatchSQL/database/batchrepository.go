package database

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

var _ TweetRepository = &BatchRepository{}

var (
	ErrBufferFull = errors.New("buffer is full")
	ErrNotFound   = errors.New("item not found")
)

type BatchRepository struct {
	db     *gorm.DB
	buffer chan inQueueSelect
}

func NewBatchRepository(db *gorm.DB, bufferSize int) *BatchRepository {
	b := &BatchRepository{db: db, buffer: make(chan inQueueSelect, bufferSize)}
	go b.selectInBatchJob()
	return b
}

func (b *BatchRepository) ByID(ctx context.Context, id uint) (Tweet, error) {
	promise := make(chan queueResponse)
	iqs := inQueueSelect{
		ID:      id,
		Promise: promise,
	}

	select {
	case b.buffer <- iqs:
	default:
		return Tweet{}, ErrBufferFull
	}

	select {
	case <-ctx.Done():
		return Tweet{}, ctx.Err()
	case result := <-promise:
		return result.tweet, result.error
	}
}

func (b *BatchRepository) selectInBatchJob() {
	for {
		<-time.After(time.Millisecond * 20)

		l := len(b.buffer)
		if l == 0 {
			continue
		}

		tweets := make(map[uint]queueResponse)
		buffer := make([]inQueueSelect, l)
		for i := 0; i < l; i++ {
			item := <-b.buffer
			buffer[i] = item
			tweets[item.ID] = queueResponse{}
		}

		var ids []uint
		for id, _ := range tweets {
			ids = append(ids, id)
		}
		var results []Tweet
		if err := b.db.Where("id IN ?", ids).Find(&results).Error; err != nil {
			for u, response := range tweets {
				response.error = err
				tweets[u] = response
			}
		} else {
			for _, result := range results {
				response := tweets[result.ID]
				response.tweet = result
				tweets[result.ID] = response
			}
		}

		for _, queueSelect := range buffer {
			response := tweets[queueSelect.ID]
			if response.error == nil && response.tweet.ID == 0 {
				response.error = ErrNotFound
			}

			queueSelect.Promise <- response
		}
	}
}

// queue definitions
type queueResponse struct {
	tweet Tweet
	error error
}

type inQueueSelect struct {
	ID      uint
	Promise chan queueResponse
}
