package cq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// cache is a generic interface that defines methods for finding items of any type.
type cache[T any] interface {
	// FindOne method takes a context and a genKey as parameters.
	// It returns a pointer to the found item of type T and a boolean value indicating whether the item was found or not, and an error if any occurred during the operation.
	FindOne(ctx context.Context, key string) (*T, bool, error)

	// FindMany method takes a context and a slice of keys as parameters.
	// It returns a slice of pointers to the found items of type T, a slice of keys that were not found, and an error if any occurred during the operation.
	FindMany(ctx context.Context, keys []string) ([]*T, []string, error)

	// SetOne method takes a context, a genKey, and a pointer to an item of type T as parameters.
	// It returns an error if any occurred during the operation.
	SetOne(ctx context.Context, key string, item *T, expiration time.Duration) error

	// SetMany method takes a context, a slice of keys, and a slice of pointers to items of type T as parameters.
	// It returns an error if any occurred during the operation.
	SetMany(ctx context.Context, keys []string, items []*T, expiration time.Duration) error
}

var _ cache[any] = (*RedisCache[any])(nil)

type RedisCache[T any] struct {
	client redis.UniversalClient
}

func NewRedisCache[T any](client redis.UniversalClient) *RedisCache[T] {
	return &RedisCache[T]{
		client: client,
	}
}

func (rc *RedisCache[T]) FindOne(ctx context.Context, key string) (*T, bool, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var item T
	err = json.Unmarshal([]byte(val), &item)
	if err != nil {
		return nil, false, err
	}

	return &item, true, nil
}

func (rc *RedisCache[T]) FindMany(ctx context.Context, keys []string) ([]*T, []string, error) {
	values, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}

	var items []*T
	var notFoundKeys []string

	for i, val := range values {
		if val == nil {
			notFoundKeys = append(notFoundKeys, keys[i])
			continue
		}
		var item T
		s, _ := val.(string)
		err = json.Unmarshal([]byte(s), &item)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, &item)
	}

	return items, notFoundKeys, nil
}

func (rc *RedisCache[T]) SetOne(ctx context.Context, key string, item *T, expiration time.Duration) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	err = rc.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (rc *RedisCache[T]) SetMany(ctx context.Context, keys []string, items []*T, expiration time.Duration) error {
	// cause MSET does not support expiration, so we have to use pipeline
	pipe := rc.client.Pipeline()
	for i, key := range keys {
		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}
		pipe.Set(ctx, key, data, expiration)
	}
	exec, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	for _, cmd := range exec {
		if cmd.Err() != nil {
			err = errors.Join(err, cmd.Err())
		}
	}

	return err
}
