package cq

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) FindOne(ctx context.Context, id int64) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepo) FindMany(ctx context.Context, ids []int64) ([]*User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*User), args.Error(1)
}

func TestFindOneReturnsUserWhenKeyExistsInCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	repo := new(MockRepo)
	cache := NewCache[User](repo, func(u *User) int64 { return int64(u.ID) }, "user", client, 0)
	ctx := context.Background()

	user := User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
	client.Set(ctx, "user:1", serialize(user, t), 0)

	result, err := cache.FindOne(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, user, *result)
}

func TestFindOneReturnsUserWhenKeyDoesNotExistInCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	repo := new(MockRepo)
	cache := NewCache[User](repo, func(u *User) int64 { return int64(u.ID) }, "user", client, 0)
	ctx := context.Background()

	client.Del(ctx, "user:1")
	user := User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
	repo.On("FindOne", ctx, int64(1)).Return(&user, nil)

	result, err := cache.FindOne(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, user, *result)
}

func TestFindManyReturnsUsersWhenKeysExistInCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	repo := new(MockRepo)
	cache := NewCache[User](repo, func(u *User) int64 { return int64(u.ID) }, "user", client, 0)
	ctx := context.Background()

	users := []User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}
	client.MSet(ctx, "user:1", serialize(users[0], t), "user:2", serialize(users[1], t))

	results, err := cache.FindMany(ctx, []int64{1, 2})

	assert.NoError(t, err)
	for i, user := range users {
		assert.Equal(t, user, *results[i])
	}
}

func TestFindManyReturnsUsersWhenKeysDoNotExistInCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	repo := new(MockRepo)
	cache := NewCache[User](repo, func(u *User) int64 { return int64(u.ID) }, "user", client, 0)
	ctx := context.Background()

	client.Del(ctx, "user:1", "user:2")
	users := []User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}
	repo.On("FindMany", ctx, []int64{1, 2}).Return([]*User{&users[0], &users[1]}, nil)

	results, err := cache.FindMany(ctx, []int64{1, 2})

	assert.NoError(t, err)
	for i, user := range users {
		assert.Equal(t, user, *results[i])
	}
}

func TestFindManyReturnsUsersWhenAllKeysExistInCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	repo := new(MockRepo)
	cache := NewCache[User](repo, func(u *User) int64 { return int64(u.ID) }, "user", client, 0)
	ctx := context.Background()

	users := []User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}
	client.MSet(ctx, "user:1", serialize(users[0], t), "user:2", serialize(users[1], t))

	results, err := cache.FindMany(ctx, []int64{1, 2})

	assert.NoError(t, err)
	for i, user := range users {
		assert.Equal(t, user, *results[i])
	}
}

func TestFindManyReturnsUsersWhenSomeKeysExistInCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	repo := new(MockRepo)
	cache := NewCache[User](repo, func(u *User) int64 { return int64(u.ID) }, "user", client, 0)
	ctx := context.Background()

	client.Del(ctx, "user:1")
	users := []User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}
	client.Set(ctx, "user:1", serialize(users[0], t), 0)
	repo.On("FindMany", ctx, []int64{2}).Return([]*User{&users[1]}, nil)

	results, err := cache.FindMany(ctx, []int64{1, 2})

	assert.NoError(t, err)
	for i, user := range users {
		assert.Equal(t, user, *results[i])
	}
}

func TestFindManyReturnsUsersWhenNoKeysExistInCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{})
	repo := new(MockRepo)
	cache := NewCache[User](repo, func(u *User) int64 { return int64(u.ID) }, "user", client, 0)
	ctx := context.Background()

	client.Del(ctx, "user:1", "user:2")
	users := []User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}
	repo.On("FindMany", ctx, []int64{1, 2}).Return([]*User{&users[0], &users[1]}, nil)

	results, err := cache.FindMany(ctx, []int64{1, 2})

	assert.NoError(t, err)
	for i, user := range users {
		assert.Equal(t, user, *results[i])
	}
}
