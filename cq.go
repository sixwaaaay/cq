package cq

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repo[T any] interface {
	FindOne(ctx context.Context, key int64) (*T, error)
	FindMany(ctx context.Context, keys []int64) ([]*T, error)
}

type Cache[T any] struct {
	repo   Repo[T]
	Id     func(*T) int64
	Prefix string
	cache  RedisCache[T]
	expire time.Duration
}

func NewCache[T any](
	repo Repo[T],
	id func(*T) int64,
	prefix string,
	client redis.UniversalClient,
	expire time.Duration,
) *Cache[T] {
	return &Cache[T]{
		repo:   repo,
		Id:     id,
		Prefix: prefix,
		cache: RedisCache[T]{
			client: client,
		},
		expire: expire,
	}
}

func genKey(prefix string, id int64) string {
	return prefix + ":" + strconv.FormatInt(id, 10)
}

func key2int64(prefix string, key string) (int64, error) {
	return strconv.ParseInt(key[len(prefix)+1:], 10, 64)
}

func (c *Cache[T]) FindOne(ctx context.Context, id int64) (*T, error) {
	key := genKey(c.Prefix, id)
	item, found, err := c.cache.FindOne(ctx, key)
	if err != nil {
		return nil, err
	}
	if found {
		return item, nil
	}
	item, err = c.repo.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	err = c.cache.SetOne(ctx, key, item, c.expire)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Cache[T]) FindMany(ctx context.Context, ids []int64) ([]*T, error) {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = genKey(c.Prefix, id)
	}
	items, notFoundKeys, err := c.cache.FindMany(ctx, keys)
	if err != nil {
		return nil, err
	}
	if len(notFoundKeys) == 0 {
		return items, nil
	}
	ids = make([]int64, len(notFoundKeys))
	for i, key := range notFoundKeys {
		ids[i], err = key2int64(c.Prefix, key)
		if err != nil {
			return nil, err
		}
	}
	missingItems, err := c.repo.FindMany(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(missingItems) == 0 {
		return items, nil
	}
	var missingKeys []string
	for _, item := range missingItems {
		key := genKey(c.Prefix, c.Id(item))
		missingKeys = append(missingKeys, key)
	}
	err = c.cache.SetMany(ctx, missingKeys, missingItems, c.expire)

	if err != nil {
		return nil, err
	}
	return append(items, missingItems...), nil
}
