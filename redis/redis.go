package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/viccon/pulse"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	redisClient *redis.Client
}

func New(addr, password string) *Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &Client{redisClient}
}

func (c *Client) Write(ctx context.Context, session pulse.CodingSession) error {
	// Check if we already have a session for this date.
	cmd := c.redisClient.Get(ctx, session.DateString())
	data, err := cmd.Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	if !errors.Is(err, redis.Nil) {
		var prevSession pulse.CodingSession
		unmarshalErr := json.Unmarshal([]byte(data), &prevSession)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		session = session.Merge(prevSession)
	}

	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return c.redisClient.Set(ctx, session.DateString(), bytes, 0).Err()
}

func (c *Client) ReadAll(ctx context.Context) (pulse.CodingSessions, error) {
	var sessions pulse.CodingSessions
	var keys []string

	iter := c.redisClient.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	batchSize := 25
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batchKeys := keys[i:end]
		cmd := c.redisClient.MGet(ctx, batchKeys...)
		data, err := cmd.Result()
		if err != nil {
			return nil, err
		}

		for _, item := range data {
			if item == nil {
				continue
			}

			var session pulse.CodingSession
			if err := json.Unmarshal([]byte(item.(string)), &session); err != nil {
				return nil, err
			}

			sessions = append(sessions, session)
		}
	}

	sort.Sort(sessions)
	return sessions, nil
}
