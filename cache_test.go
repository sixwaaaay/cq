package cq

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func serialize[T any](item T, t *testing.T) []byte {
	bytes, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func deserialize[T any](bytes []byte, t *testing.T) *T {
	var item T
	err := json.Unmarshal(bytes, &item)
	if err != nil {
		t.Fatal(err)
	}
	return &item
}

func TestFindOneReturnsUserWhenKeyExists(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	cache := NewRedisCache[User](client)
	ctx := context.Background()

	user := User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
	client.Set(ctx, "user:1", serialize(user, t), 0)

	result, found, err := cache.FindOne(ctx, "user:1")

	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, user, *result)
}

func TestFindManyReturnsUsersWhenKeysExist(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	cache := NewRedisCache[User](client)
	ctx := context.Background()

	users := []User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}
	client.MSet(ctx, "user:1", serialize(users[0], t), "user:2", serialize(users[1], t))

	results, notFoundKeys, err := cache.FindMany(ctx, []string{"user:1", "user:2"})

	assert.NoError(t, err)
	assert.Empty(t, notFoundKeys)
	//assert.Equal(t, users, results)
	for i, user := range users {
		assert.Equal(t, user, *results[i])
	}
}

func TestSetOneStoresUser(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	cache := NewRedisCache[User](client)
	ctx := context.Background()

	user := User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
	err := cache.SetOne(ctx, "user:1", &user, 0)

	assert.NoError(t, err)

	result, err := client.Get(ctx, "user:1").Result()
	assert.NoError(t, err)
	//assert.Equal(t, user, result)
	assert.Equal(t, user, *deserialize[User]([]byte(result), t))
}

func TestSetManyStoresUsers(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	cache := NewRedisCache[User](client)
	ctx := context.Background()

	users := []*User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}
	err := cache.SetMany(ctx, []string{"user:1", "user:2"}, users, 0)

	assert.NoError(t, err)

	result1, err1 := client.Get(ctx, "user:1").Result()
	result2, err2 := client.Get(ctx, "user:2").Result()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	//assert.Equal(t, users[0], result1)
	//assert.Equal(t, users[1], result2)
	assert.Equal(t, users[0], deserialize[User]([]byte(result1), t))
	assert.Equal(t, users[1], deserialize[User]([]byte(result2), t))
}
