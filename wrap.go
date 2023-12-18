package cq

import "context"

type wrapRepo[T any] struct {
	findOne  func(ctx context.Context, id int64) (*T, error)
	findMany func(ctx context.Context, ids []int64) ([]T, error)
}

func NewWrapRepo[T any](findOne func(ctx context.Context, id int64) (*T, error), findMany func(ctx context.Context, ids []int64) ([]T, error)) wrapRepo[T] {
	return wrapRepo[T]{
		findOne:  findOne,
		findMany: findMany,
	}
}

func (r wrapRepo[T]) FindOne(ctx context.Context, id int64) (*T, error) {
	return r.findOne(ctx, id)
}

func (r wrapRepo[T]) FindMany(ctx context.Context, ids []int64) ([]T, error) {
	return r.findMany(ctx, ids)
}
