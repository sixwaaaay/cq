package cq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache[T any] struct {
	cli    redis.UniversalClient
	expire time.Duration
}

type Option struct {
	Expire time.Duration
	Client redis.UniversalClient
}

// NewCache returns a new Cache[T] instance.
func NewCache[T any](option Option) *Cache[T] {
	c := Cache[T]{
		cli:    option.Client,
		expire: option.Expire,
	}
	c.expire = time.Minute
	return &c
}

// FindOne returns a function that can be used to find a single item.
func (c *Cache[T]) FindOne(name string, fn func(ctx context.Context, id int64) (*T, error)) func(ctx context.Context, id int64) (*T, error) {
	return func(ctx context.Context, id int64) (*T, error) {
		key := fmt.Sprintf("%s:%d", name, id)
		val, err := c.cli.Get(ctx, key).Result()
		if err == redis.Nil {
			t, err := fn(ctx, id)
			if err != nil && t == nil {
				return nil, errors.New("not found")
			}
			b, err := json.Marshal(t)
			if err != nil {
				return nil, err
			}
			err = c.cli.Set(ctx, key, b, c.expire).Err()
			if err != nil {
				return nil, err
			}
			return t, nil
		} else if err != nil {
			return fn(ctx, id)
		}
		var t T
		if err = json.Unmarshal([]byte(val), &t); err != nil {
			return nil, err
		}
		return &t, nil
	}
}

// FindMany returns a function that can be used to find multiple items.
func (c *Cache[T]) FindMany(name string, mapping func(*T) int64, fn func(ctx context.Context, ids []int64) ([]*T, error)) func(ctx context.Context, ids []int64) ([]*T, error) {
	return func(ctx context.Context, ids []int64) ([]*T, error) {
		var ts []*T
		var missIds []int64
		for _, id := range ids {
			key := fmt.Sprintf("%s:%d", name, id)
			val, err := c.cli.Get(ctx, key).Result()
			if err == redis.Nil || val == "" || err != nil {
				missIds = append(missIds, id)
				continue
			}
			var t T
			if err = json.Unmarshal([]byte(val), &t); err != nil {
				return nil, err
			}
			ts = append(ts, &t)
		}
		if len(missIds) > 0 {
			missTs, err := fn(ctx, missIds)
			if err != nil {
				return nil, err
			}
			for _, t := range missTs {
				b, err := json.Marshal(t)
				if err != nil {
					return nil, err
				}
				key := fmt.Sprintf("%s:%d", name, mapping(t))
				err = c.cli.Set(ctx, key, b, c.expire).Err()
				if err != nil {
					return nil, err
				}
				ts = append(ts, t)
			}
		}
		return ts, nil
	}
}
