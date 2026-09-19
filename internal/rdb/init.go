package rdb

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

type DB interface {
	SetSession(ctx context.Context, sessionKey, userID string) error
	GetKey(ctx context.Context, device string) (string, error)
	Close() error
}

func NewClient(redisAddr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	rdb := &Client{
		rdb: redisClient,
	}

	return rdb, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
